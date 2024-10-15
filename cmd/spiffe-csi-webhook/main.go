package main

import (
    "fmt"
    "time"
    "log"
    "io/ioutil"
    "net/http"
    "encoding/json"

    admissionv1 "k8s.io/api/admission/v1"
    corev1 "k8s.io/api/core/v1"
    appsv1 "k8s.io/api/apps/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
    scheme  =   runtime.NewScheme()
    codecs  =   serializer.NewCodecFactory(scheme)
)

func main() {
    mux := http.NewServeMux()

	mux.HandleFunc("/mutate", handleMutate)

	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}

    log.Println("Starting SPIFFE CSI webhook server on :8443...")
	log.Fatal(s.ListenAndServeTLS("/ssl/spiffe-csi-webhook.pem", "/ssl/spiffe-csi-webhook.key"))
}

func handleMutate(w http.ResponseWriter, r *http.Request) {
    var body []byte
    if r.Body != nil {
        if data, err := ioutil.ReadAll(r.Body); err == nil {
            body = data
        }
    }

    if len(body) == 0 {
        http.Error(w, "empty body", http.StatusBadRequest)
        return
    }

    contentType := r.Header.Get("Content-Type")
    if contentType != "application/json" {
        http.Error(w, fmt.Sprintf("contentType=%s, expect application/json", contentType), http.StatusBadRequest)
        return
    }

    review := admissionv1.AdmissionReview{}
    deserializer := codecs.UniversalDeserializer()
    if _, _, err := deserializer.Decode(body, nil, &review); err != nil {
        http.Error(w, fmt.Sprintf("could not decode body: %v", err), http.StatusBadRequest)
        return
    }

    if review.Request.Kind.Kind != "Deployment" {
        http.Error(w, "only supports deployments", http.StatusBadRequest)
        return
    }

    deployment := appsv1.Deployment{}
    if err := json.Unmarshal(review.Request.Object.Raw, &deployment); err != nil {
        http.Error(w, fmt.Sprintf("could not unmarshal deployment: %v", err), http.StatusBadRequest)
        return
    }

    annotations := deployment.Annotations
    if annotations == nil {
        annotations = make(map[string]string)
    }

    if val, exists := annotations["spiffe.io/inject-cert"]; exists && val == "true" {
        fmt.Println("Found a deployment that requires injection of SPIRE certificates")

        writeResponse(w, &review, nil)
        return
    }

    // Add the ephemeral volume (emptyDir) to the deployment spec
	volume := corev1.Volume{
		Name: "ephemeral-volume",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

    // Add the volume mount to each container
    for i := range deployment.Spec.Template.Spec.Containers {
		deployment.Spec.Template.Spec.Containers[i].VolumeMounts = append(
			deployment.Spec.Template.Spec.Containers[i].VolumeMounts,
			corev1.VolumeMount{
				Name:      "ephemeral-volume",
				MountPath: "/mnt/ephemeral",
			},
		)
	}

   	// Add the volume to the pod spec
	deployment.Spec.Template.Spec.Volumes = append(
		deployment.Spec.Template.Spec.Volumes,
		volume,
	)

    // Create the patch to modify the deployment
	patchBytes, err := json.Marshal([]map[string]interface{}{
		{
			"op":    "replace",
			"path":  "/spec/template/spec/volumes",
			"value": deployment.Spec.Template.Spec.Volumes,
		},
		{
			"op":    "replace",
			"path":  "/spec/template/spec/containers",
			"value": deployment.Spec.Template.Spec.Containers,
		},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("could not create patch: %v", err), http.StatusInternalServerError)
		return
	}

    writeResponse(w, &review, patchBytes)
}

// Helper function to write the admission response
func writeResponse(w http.ResponseWriter, review *admissionv1.AdmissionReview, patch []byte) {
	response := admissionv1.AdmissionResponse{
		Allowed: true,
		UID:     review.Request.UID,
	}

	if patch != nil {
		response.Patch = patch
		response.PatchType = func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}()
	}

	review.Response = &response

	respBytes, err := json.Marshal(review)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

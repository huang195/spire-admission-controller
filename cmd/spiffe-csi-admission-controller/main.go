package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"

    admissionv1 "k8s.io/api/admission/v1"
    appsv1 "k8s.io/api/apps/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
    scheme  =   runtime.NewScheme()
    codecs  =   serializer.NewCodecFactory(scheme)
)

func main() {
    http.HandleFunc("/mutate", handleMutate)
    fmt.Println("Starting SPIFFE CSI webhook server on :8080...")
    http.ListenAndServe(":8080", nil)
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
        http.Error(w, fmt.Sprintf("contentType=%s, expect application/json", contentType)
        return
    }

    review := admissionv1.AdmissionReview{}
    deserializer := codecs.UniversalDeserializer()
    if _, _, err := deserializer.Decode(body, nil &review); err != nil {
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
    }
}

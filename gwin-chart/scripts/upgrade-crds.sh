#!/bin/bash
set -e

NAMESPACE=$1
CONTROLLER_NAME=$2
GATEWAY_API_CRD_FILE=/crds/resources/gateway-api.yaml
VERSION_TO_BUMP="v0.6.2"

echo "Applying gwin.yandex.cloud_gatewaypolicies.yaml"
kubectl apply -f /crds/policies/gwin.yandex.cloud_gatewaypolicies.yaml

echo "Applying gwin.yandex.cloud_routepolicies.yaml"
kubectl apply -f /crds/policies/gwin.yandex.cloud_routepolicies.yaml

echo "Applying gwin.yandex.cloud_servicepolicies.yaml"
kubectl apply -f /crds/policies/gwin.yandex.cloud_servicepolicies.yaml

echo "Applying gwin.yandex.cloud_ingresspolicies.yaml"
kubectl apply -f /crds/policies/gwin.yandex.cloud_ingresspolicies.yaml

echo "Applying gwin.yandex.cloud_yccertificates.yaml"
kubectl apply -f /crds/resources/gwin.yandex.cloud_yccertificates.yaml

echo "Applying gwin.yandex.cloud_ycstoragebuckets.yaml"
kubectl apply -f /crds/resources/gwin.yandex.cloud_ycstoragebuckets.yaml

echo "Applying gwin.yandex.cloud_ingressbackendgroups.yaml"
kubectl apply -f /crds/resources/gwin.yandex.cloud_ingressbackendgroups.yaml

get_crd_bundle_version() {
    local crd_name=$1
    kubectl get crd "$crd_name" -o json | jq -r '.metadata.annotations."gateway.networking.k8s.io/bundle-version"'
}
export -f get_crd_bundle_version

deployment_exists() {
    local namespace=$1
    local deployment_name=$2
    kubectl get deployment "$deployment_name" -n "$namespace" >/dev/null 2>&1
}

echo "Checking CRD bundle-version"
min_bundle_version=$(kubectl get crd -A | grep gateway.networking.k8s.io | awk '{print $1}' | xargs -I {} bash -c "get_crd_bundle_version {}" | sort | head -1)
if [[ "$min_bundle_version" > "$VERSION_TO_BUMP" ]]; then
    echo "CRD bundle-version is higher than $VERSION_TO_BUMP: $min_bundle_version, upgrade not required"
    exit 0
fi
echo "Found CRD bundle-version: $min_bundle_version, upgrade required"

echo "Checking existing grpcroutes and referencegrants"
grpcroutes_version=$(get_crd_bundle_version grpcroutes.gateway.networking.k8s.io)
grpcroutes_count=$(kubectl get grpcroutes.gateway.networking.k8s.io --all-namespaces --no-headers 2>/dev/null | wc -l)
echo "Found $grpcroutes_count grpcroutes with bundle-version: $grpcroutes_version"
referencegrants_version=$(get_crd_bundle_version referencegrants.gateway.networking.k8s.io)
referencegrants_count=$(kubectl get referencegrants.gateway.networking.k8s.io --all-namespaces --no-headers 2>/dev/null | wc -l)
echo "Found $referencegrants_count referencegrants with bundle-version: $referencegrants_version"
if { [ "$grpcroutes_count" -gt 0 ] && [ "$grpcroutes_version" -le "$VERSION_TO_BUMP" ]; } || \
    { [ "$referencegrants_count" -gt 0 ] && [ "$referencegrants_version" -le "$VERSION_TO_BUMP" ]; } then
    echo "There are existing grpcroutes or referencegrants present. Can't perform crds auto upgrade."
    exit 1
fi
echo "No existing grpcroutes or referencegrants found"

if deployment_exists "$NAMESPACE" "$CONTROLLER_NAME"; then
    echo "Scaling down $CONTROLLER_NAME to 0 replicas"
    kubectl scale deployment "$CONTROLLER_NAME" --replicas=0 -n "$NAMESPACE"
    while [[ $(kubectl get pods -l app="$CONTROLLER_NAME" -n "$NAMESPACE" --field-selector=status.phase=Running | wc -l) -gt 0 ]]; do
        echo "Waiting for pods to terminate..."
        sleep 1
    done
    echo "Scaling down complete"
else
    echo "Deployment $CONTROLLER_NAME not found in namespace $NAMESPACE. Skipping scaling operations."
fi

echo "Deleting conflicting grpcroutes and referencegrants CRDs..."
if [ "$grpcroutes_count" -eq 0 ]; then
    kubectl delete crd grpcroutes.gateway.networking.k8s.io
fi
if [ "$referencegrants_count" -eq 0 ]; then
    kubectl delete crd referencegrants.gateway.networking.k8s.io
fi

echo "Applying new CRDs"
kubectl apply -f "$GATEWAY_API_CRD_FILE"
echo "Upgrade complete"

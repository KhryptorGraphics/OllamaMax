#!/bin/bash
# Validation script for ELK Stack and Jaeger deployment
# Comment 6 Implementation Verification

set -e

NAMESPACE="ollamamax-monitoring"
TIMEOUT=300

echo "======================================================"
echo "ELK Stack and Jaeger Deployment Validation"
echo "======================================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if a resource exists
check_resource() {
    local resource_type=$1
    local resource_name=$2
    local namespace=$3

    if kubectl get "$resource_type" "$resource_name" -n "$namespace" &> /dev/null; then
        echo -e "${GREEN}✓${NC} $resource_type/$resource_name exists"
        return 0
    else
        echo -e "${RED}✗${NC} $resource_type/$resource_name NOT FOUND"
        return 1
    fi
}

# Function to check pod status
check_pod_status() {
    local app_label=$1
    local namespace=$2
    local expected_count=$3

    local ready_count=$(kubectl get pods -n "$namespace" -l "app=$app_label" --field-selector=status.phase=Running 2>/dev/null | grep -c "Running" || echo 0)

    if [ "$ready_count" -ge "$expected_count" ]; then
        echo -e "${GREEN}✓${NC} $app_label: $ready_count/$expected_count pods running"
        return 0
    else
        echo -e "${YELLOW}⚠${NC} $app_label: $ready_count/$expected_count pods running (waiting...)"
        return 1
    fi
}

# Check namespace
echo "1. Checking namespace..."
check_resource namespace "$NAMESPACE" "" || exit 1
echo ""

# Check Elasticsearch
echo "2. Checking Elasticsearch StatefulSet..."
check_resource statefulset elasticsearch "$NAMESPACE"
check_resource service elasticsearch "$NAMESPACE"
check_resource service elasticsearch-http "$NAMESPACE"

echo "   Checking Elasticsearch pods..."
check_pod_status elasticsearch "$NAMESPACE" 3
echo ""

# Check Logstash
echo "3. Checking Logstash Deployment..."
check_resource deployment logstash "$NAMESPACE"
check_resource service logstash "$NAMESPACE"
check_resource configmap logstash-pipeline "$NAMESPACE"

echo "   Checking Logstash pods..."
check_pod_status logstash "$NAMESPACE" 1
echo ""

# Check Kibana
echo "4. Checking Kibana Deployment..."
check_resource deployment kibana "$NAMESPACE"
check_resource service kibana "$NAMESPACE"

echo "   Checking Kibana pods..."
check_pod_status kibana "$NAMESPACE" 1
echo ""

# Check Filebeat
echo "5. Checking Filebeat DaemonSet..."
check_resource daemonset filebeat "$NAMESPACE"
check_resource serviceaccount filebeat "$NAMESPACE"
check_resource clusterrole filebeat ""
check_resource clusterrolebinding filebeat ""
check_resource configmap filebeat-config "$NAMESPACE"

echo "   Checking Filebeat pods..."
NODE_COUNT=$(kubectl get nodes --no-headers 2>/dev/null | wc -l)
check_pod_status filebeat "$NAMESPACE" 1
echo ""

# Check Jaeger
echo "6. Checking Jaeger Deployment..."
check_resource deployment jaeger "$NAMESPACE"
check_resource service jaeger "$NAMESPACE"
check_resource service jaeger-ui "$NAMESPACE"

echo "   Checking Jaeger pods..."
check_pod_status jaeger "$NAMESPACE" 1
echo ""

# Test Elasticsearch cluster health
echo "7. Testing Elasticsearch cluster health..."
if kubectl get pods -n "$NAMESPACE" -l app=elasticsearch | grep -q "Running"; then
    ES_POD=$(kubectl get pods -n "$NAMESPACE" -l app=elasticsearch -o jsonpath='{.items[0].metadata.name}')
    echo "   Using pod: $ES_POD"

    HEALTH=$(kubectl exec -n "$NAMESPACE" "$ES_POD" -- curl -s http://localhost:9200/_cluster/health 2>/dev/null || echo "ERROR")

    if [[ "$HEALTH" == *"\"status\":\"green\""* ]] || [[ "$HEALTH" == *"\"status\":\"yellow\""* ]]; then
        echo -e "${GREEN}✓${NC} Elasticsearch cluster is healthy"
        echo "   Status: $(echo "$HEALTH" | grep -o '"status":"[^"]*"')"
        echo "   Nodes: $(echo "$HEALTH" | grep -o '"number_of_nodes":[0-9]*')"
    else
        echo -e "${YELLOW}⚠${NC} Elasticsearch cluster health check failed or not ready yet"
    fi
else
    echo -e "${YELLOW}⚠${NC} No running Elasticsearch pods found"
fi
echo ""

# Test Logstash connectivity
echo "8. Testing Logstash connectivity..."
if kubectl get pods -n "$NAMESPACE" -l app=logstash | grep -q "Running"; then
    LOGSTASH_POD=$(kubectl get pods -n "$NAMESPACE" -l app=logstash -o jsonpath='{.items[0].metadata.name}')
    echo "   Using pod: $LOGSTASH_POD"

    LOGSTASH_STATUS=$(kubectl exec -n "$NAMESPACE" "$LOGSTASH_POD" -- curl -s http://localhost:9600 2>/dev/null || echo "ERROR")

    if [[ "$LOGSTASH_STATUS" == *"version"* ]]; then
        echo -e "${GREEN}✓${NC} Logstash is responding"
    else
        echo -e "${YELLOW}⚠${NC} Logstash health check failed or not ready yet"
    fi
else
    echo -e "${YELLOW}⚠${NC} No running Logstash pods found"
fi
echo ""

# Test Jaeger UI
echo "9. Testing Jaeger UI..."
if kubectl get pods -n "$NAMESPACE" -l app=jaeger | grep -q "Running"; then
    JAEGER_POD=$(kubectl get pods -n "$NAMESPACE" -l app=jaeger -o jsonpath='{.items[0].metadata.name}')
    echo "   Using pod: $JAEGER_POD"

    JAEGER_STATUS=$(kubectl exec -n "$NAMESPACE" "$JAEGER_POD" -- curl -s http://localhost:16686 2>/dev/null || echo "ERROR")

    if [[ "$JAEGER_STATUS" == *"Jaeger UI"* ]] || [[ "$JAEGER_STATUS" == *"<!doctype html>"* ]]; then
        echo -e "${GREEN}✓${NC} Jaeger UI is responding"
    else
        echo -e "${YELLOW}⚠${NC} Jaeger UI health check failed or not ready yet"
    fi
else
    echo -e "${YELLOW}⚠${NC} No running Jaeger pods found"
fi
echo ""

# Check Elasticsearch indices
echo "10. Checking Elasticsearch indices..."
if kubectl get pods -n "$NAMESPACE" -l app=elasticsearch | grep -q "Running"; then
    ES_POD=$(kubectl get pods -n "$NAMESPACE" -l app=elasticsearch -o jsonpath='{.items[0].metadata.name}')

    INDICES=$(kubectl exec -n "$NAMESPACE" "$ES_POD" -- curl -s "http://localhost:9200/_cat/indices?v" 2>/dev/null || echo "ERROR")

    if [[ "$INDICES" == *"ollamamax-logs"* ]]; then
        echo -e "${GREEN}✓${NC} OllamaMax log indices found"
        echo "$INDICES" | grep "ollamamax-logs"
    elif [[ "$INDICES" != "ERROR" ]]; then
        echo -e "${YELLOW}⚠${NC} No OllamaMax log indices yet (logs may not have been ingested)"
        echo "   Available indices:"
        echo "$INDICES" | head -5
    else
        echo -e "${YELLOW}⚠${NC} Cannot retrieve indices (Elasticsearch may not be ready)"
    fi
else
    echo -e "${YELLOW}⚠${NC} No running Elasticsearch pods found"
fi
echo ""

# Display service endpoints
echo "11. Service Endpoints..."
echo ""
echo "   Kibana UI:"
kubectl get svc -n "$NAMESPACE" kibana -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null && echo ":5601" || echo "      (LoadBalancer pending or use port-forward: kubectl port-forward -n $NAMESPACE svc/kibana 5601:5601)"
echo ""
echo "   Jaeger UI:"
kubectl get svc -n "$NAMESPACE" jaeger-ui -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null && echo ":16686" || echo "      (LoadBalancer pending or use port-forward: kubectl port-forward -n $NAMESPACE svc/jaeger-ui 16686:16686)"
echo ""
echo "   Elasticsearch API:"
echo "      http://elasticsearch-http.$NAMESPACE.svc.cluster.local:9200 (internal)"
echo "      Port-forward: kubectl port-forward -n $NAMESPACE svc/elasticsearch-http 9200:9200"
echo ""

# Summary
echo "======================================================"
echo "Validation Summary"
echo "======================================================"
echo ""

ALL_READY=true

# Count pods
ES_READY=$(kubectl get pods -n "$NAMESPACE" -l app=elasticsearch --field-selector=status.phase=Running 2>/dev/null | grep -c "Running" || echo 0)
LOGSTASH_READY=$(kubectl get pods -n "$NAMESPACE" -l app=logstash --field-selector=status.phase=Running 2>/dev/null | grep -c "Running" || echo 0)
KIBANA_READY=$(kubectl get pods -n "$NAMESPACE" -l app=kibana --field-selector=status.phase=Running 2>/dev/null | grep -c "Running" || echo 0)
FILEBEAT_READY=$(kubectl get pods -n "$NAMESPACE" -l app=filebeat --field-selector=status.phase=Running 2>/dev/null | grep -c "Running" || echo 0)
JAEGER_READY=$(kubectl get pods -n "$NAMESPACE" -l app=jaeger --field-selector=status.phase=Running 2>/dev/null | grep -c "Running" || echo 0)

echo "Component Status:"
echo "  Elasticsearch: $ES_READY/3 pods"
echo "  Logstash:      $LOGSTASH_READY/1 pods"
echo "  Kibana:        $KIBANA_READY/1 pods"
echo "  Filebeat:      $FILEBEAT_READY/$NODE_COUNT DaemonSet pods"
echo "  Jaeger:        $JAEGER_READY/1 pods"
echo ""

if [ "$ES_READY" -ge 3 ] && [ "$LOGSTASH_READY" -ge 1 ] && [ "$KIBANA_READY" -ge 1 ] && [ "$FILEBEAT_READY" -ge 1 ] && [ "$JAEGER_READY" -ge 1 ]; then
    echo -e "${GREEN}✓ All core components are running${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Access Kibana UI to view logs"
    echo "  2. Access Jaeger UI to view traces"
    echo "  3. Configure index patterns in Kibana (ollamamax-logs-*)"
    echo "  4. Instrument applications to send traces to Jaeger"
else
    echo -e "${YELLOW}⚠ Some components are not fully ready yet${NC}"
    echo ""
    echo "Wait for all pods to be ready, then run this script again."
    echo ""
    echo "To monitor pod status:"
    echo "  kubectl get pods -n $NAMESPACE -w"
fi

echo ""
echo "======================================================"
echo "Validation complete"
echo "======================================================"

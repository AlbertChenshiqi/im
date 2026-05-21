#!/usr/bin/env bash
# 按模块生成 K8s 清单：deploy/k8s/overlays/local/apps/<module>/
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/deploy/k8s/overlays/local/apps"
CFG="$ROOT/deploy/k8s/overlays/local/config"
rm -rf "$OUT"
mkdir -p "$OUT"

TAG="${IMAGE_TAG:-dev}"
GATEWAY_REPLICAS="${GATEWAY_REPLICAS:-2}"

# kind|image|bin|port|nodePort|module|configFile|probePath
SERVICES=(
  "api|im/gateway/gateway-api|gateway-api|10000|30000|gateway|gateway-api.yaml|/v1/health"
  "api|im/user/user-api|user-api|10100|30100|user|user-api.yaml|"
  "api|im/friend/friend-api|friend-api|10200|30200|friend|friend-api.yaml|"
  "api|im/group/group-api|group-api|10300|30300|group|group-api.yaml|"
  "api|im/conversation/conversation-api|conversation-api|10400|30400|conversation|conversation-api.yaml|"
  "api|im/message/message-api|message-api|10500|30500|message|message-api.yaml|"
  "api|im/notification/notification-api|notification-api|10600|30600|notification|notification-api.yaml|"
  "api|im/push/push-api|push-api|10700|30700|push|push-api.yaml|"
  "rpc|im/user/user-rpc|user-rpc|20100||user|user-rpc.yaml|"
  "rpc|im/friend/friend-rpc|friend-rpc|20200||friend|friend-rpc.yaml|"
  "rpc|im/group/group-rpc|group-rpc|20300||group|group-rpc.yaml|"
  "rpc|im/conversation/conversation-rpc|conversation-rpc|20400||conversation|conversation-rpc.yaml|"
  "rpc|im/message/message-rpc|message-rpc|20500||message|message-rpc.yaml|"
  "rpc|im/notification/notification-rpc|notification-rpc|20600||notification|notification-rpc.yaml|"
  "rpc|im/push/push-rpc|push-rpc|20700||push|push-rpc.yaml|"
  "api|im/cron/cron|cron|10800|30800|cron|cron.yaml|/health"
)

module_dir() {
  echo "$OUT/$1"
}

write_module_kustomizations() {
  local module dir
  for module in gateway user friend group conversation message notification push cron; do
    dir=$(module_dir "$module")
    [[ -d "$dir" ]] || continue
    {
      echo "apiVersion: kustomize.config.k8s.io/v1beta1"
      echo "kind: Kustomization"
      echo "commonLabels:"
      echo "  app.kubernetes.io/part-of: im"
      echo "  app.kubernetes.io/component: ${module}"
      echo "resources:"
      for f in "$dir"/*.yaml; do
        [[ -f "$f" ]] || continue
        base=$(basename "$f")
        [[ "$base" == "kustomization.yaml" ]] && continue
        echo "  - ${base}"
      done | sort
    } >"$dir/kustomization.yaml"
  done
}

emit_configmap() {
  local bin=$1 module=$2 file=$3
  local dir
  dir=$(module_dir "$module")
  mkdir -p "$dir"
  local path="$dir/${bin}-configmap.yaml"
  {
    echo "apiVersion: v1"
    echo "kind: ConfigMap"
    echo "metadata:"
    echo "  name: ${bin}-config"
    echo "  labels:"
    echo "    app.kubernetes.io/name: ${bin}"
    echo "    app.kubernetes.io/part-of: im"
    echo "    app.kubernetes.io/component: ${module}"
    echo "data:"
    echo "  ${file}: |"
    sed 's/^/    /' "$CFG/$module/$file"
  } >"$path"
}

emit_deployment() {
  local kind=$1 image=$2 bin=$3 port=$4 module=$5 file=$6 probe=$7
  local dir replicas=1
  dir=$(module_dir "$module")
  mkdir -p "$dir"
  local path="$dir/${bin}-deployment.yaml"
  local mount_path="/etc/im/${module}"

  if [[ "$bin" == "gateway-api" ]]; then
    replicas=$GATEWAY_REPLICAS
  fi

  {
    echo "apiVersion: apps/v1"
    echo "kind: Deployment"
    echo "metadata:"
    echo "  name: ${bin}"
    echo "  labels:"
    echo "    app: ${bin}"
    echo "    app.kubernetes.io/name: ${bin}"
    echo "    app.kubernetes.io/part-of: im"
    echo "    app.kubernetes.io/component: ${module}"
    echo "spec:"
    echo "  replicas: ${replicas}"
    echo "  selector:"
    echo "    matchLabels:"
    echo "      app: ${bin}"
    echo "  template:"
    echo "    metadata:"
    echo "      labels:"
    echo "        app: ${bin}"
    echo "        app.kubernetes.io/name: ${bin}"
    echo "        app.kubernetes.io/component: ${module}"
    echo "    spec:"
  } >"$path"

  if [[ "$bin" == "gateway-api" && "$replicas" -gt 1 ]]; then
    {
      echo "      affinity:"
      echo "        podAntiAffinity:"
      echo "          preferredDuringSchedulingIgnoredDuringExecution:"
      echo "          - weight: 100"
      echo "            podAffinityTerm:"
      echo "              labelSelector:"
      echo "                matchExpressions:"
      echo "                - key: app"
      echo "                  operator: In"
      echo "                  values:"
      echo "                  - gateway-api"
      echo "              topologyKey: kubernetes.io/hostname"
    } >>"$path"
  fi

  {
    echo "      containers:"
    echo "      - name: ${bin}"
    echo "        image: ${image}:${TAG}"
    echo "        imagePullPolicy: IfNotPresent"
    echo "        args: [\"-f\", \"${mount_path}/${file}\"]"
    echo "        ports:"
    echo "        - containerPort: ${port}"
    echo "          name: http"
  } >>"$path"

  if [[ "$bin" == "gateway-api" ]]; then
    {
      echo "        env:"
      echo "        - name: GATEWAY_INSTANCE_ID"
      echo "          valueFrom:"
      echo "            fieldRef:"
      echo "              fieldPath: metadata.name"
    } >>"$path"
  fi

  if [[ -n "$probe" ]]; then
    {
      echo "        livenessProbe:"
      echo "          httpGet:"
      echo "            path: ${probe}"
      echo "            port: ${port}"
      echo "          initialDelaySeconds: 10"
      echo "          periodSeconds: 15"
      echo "        readinessProbe:"
      echo "          httpGet:"
      echo "            path: ${probe}"
      echo "            port: ${port}"
      echo "          initialDelaySeconds: 5"
      echo "          periodSeconds: 10"
    } >>"$path"
  else
    {
      echo "        readinessProbe:"
      echo "          tcpSocket:"
      echo "            port: ${port}"
      echo "          initialDelaySeconds: 5"
      echo "          periodSeconds: 10"
    } >>"$path"
  fi

  {
    echo "        resources:"
    echo "          requests:"
    echo "            cpu: 50m"
    echo "            memory: 64Mi"
    echo "          limits:"
    echo "            memory: 512Mi"
    echo "        volumeMounts:"
    echo "        - name: config"
    echo "          mountPath: ${mount_path}"
    echo "          readOnly: true"
    echo "      volumes:"
    echo "      - name: config"
    echo "        configMap:"
    echo "          name: ${bin}-config"
  } >>"$path"

}

emit_service() {
  local bin=$1 port=$2 node_port=$3 module=$4
  local dir
  dir=$(module_dir "$module")
  local path="$dir/${bin}-service.yaml"
  {
    echo "apiVersion: v1"
    echo "kind: Service"
    echo "metadata:"
    echo "  name: ${bin}"
    echo "  labels:"
    echo "    app: ${bin}"
    echo "    app.kubernetes.io/name: ${bin}"
    echo "    app.kubernetes.io/component: ${module}"
    echo "spec:"
    echo "  selector:"
    echo "    app: ${bin}"
  } >"$path"

  if [[ "$bin" == "gateway-api" ]]; then
    {
      echo "  sessionAffinity: ClientIP"
      echo "  sessionAffinityConfig:"
      echo "    clientIP:"
      echo "      timeoutSeconds: 10800"
    } >>"$path"
  fi

  if [[ -n "$node_port" ]]; then
    {
      echo "  type: NodePort"
      echo "  ports:"
      echo "  - port: ${port}"
      echo "    targetPort: ${port}"
      echo "    nodePort: ${node_port}"
      echo "    name: http"
    } >>"$path"
  else
    {
      echo "  ports:"
      echo "  - port: ${port}"
      echo "    targetPort: ${port}"
      echo "    name: http"
    } >>"$path"
  fi

}

for line in "${SERVICES[@]}"; do
  IFS='|' read -r kind image bin port node_port module file probe <<<"$line"
  emit_configmap "$bin" "$module" "$file"
  emit_deployment "$kind" "$image" "$bin" "$port" "$module" "$file" "$probe"
  emit_service "$bin" "$port" "$node_port" "$module"
done

write_module_kustomizations

echo "generated apps/ per module under $OUT (gateway replicas=${GATEWAY_REPLICAS})"

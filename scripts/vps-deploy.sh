#!/usr/bin/env bash
set -euo pipefail

compose_dir="${COMPOSE_DIR:-/opt/twisky}"
compose_file="${compose_dir}/docker-compose.deploy.yml"

if [[ ! -f "${compose_file}" ]]; then
    echo "Missing ${compose_file}" >&2
    exit 1
fi

if ! docker network inspect edge >/dev/null 2>&1; then
    echo "Docker network 'edge' is missing." >&2
    echo "Bring up the shared edge stack first (pet-cpp docker/edge), then re-run." >&2
    exit 1
fi

if [[ -z "${DEPLOY_PUBLIC_BASE_URL:-}" ]]; then
    echo "Set DEPLOY_PUBLIC_BASE_URL for TWISKY_PUBLIC_BASE_URL." >&2
    exit 1
fi
export TWISKY_PUBLIC_BASE_URL="${DEPLOY_PUBLIC_BASE_URL}"

if [[ -n "${GHCR_TOKEN:-}" ]]; then
    if [[ -z "${GHCR_USER:-}" ]]; then
        echo "GHCR_USER is required when GHCR_TOKEN is set." >&2
        exit 1
    fi
    printf '%s' "${GHCR_TOKEN}" | docker login ghcr.io -u "${GHCR_USER}" --password-stdin
fi

cd "${compose_dir}"

compose() {
    docker compose -f docker-compose.deploy.yml "$@"
}

logout_ghcr() {
    if [[ -n "${GHCR_TOKEN:-}" ]]; then
        docker logout ghcr.io >/dev/null 2>&1 || true
    fi
}

compose pull
compose up -d --remove-orphans
compose ps

echo "Waiting for app /healthz..."
for attempt in $(seq 1 30); do
    if compose exec -T app wget -q -T 5 -O- "http://127.0.0.1:8080/healthz" >/dev/null 2>&1; then
        echo "App /healthz OK"
        logout_ghcr
        exit 0
    fi
    sleep 2
done

echo "App /healthz did not return 200 within 60s" >&2
compose logs --tail=50 app >&2
logout_ghcr
exit 1

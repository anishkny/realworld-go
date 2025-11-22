#!/usr/bin/env bash
set -euxo pipefail

PORT=3000
TIMEOUT=30000   # in milliseconds

npm run stop
npm run start &
npx wait-port http://localhost:${PORT} --output dots --timeout=${TIMEOUT}
npm run test:only
npm run stop


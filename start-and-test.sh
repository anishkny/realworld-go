#!/usr/bin/env bash
set -euxo pipefail

PORT=3000
URL=http://localhost:${PORT}/api
TIMEOUT=30000 # in milliseconds

npm run stop
npm run build
npm run start:only &
npx wait-port ${URL} --output dots --timeout=${TIMEOUT}
npm run test:only
npm run stop
npm run coverage

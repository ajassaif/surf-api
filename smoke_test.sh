#!/usr/bin/env bash
# End-to-end check against a running server on :8000.
set -euo pipefail
BASE=http://localhost:8000
pass=0
check() { # name expected_status method path [body]
  local name=$1 want=$2 method=$3 path=$4 body=${5:-}
  local out code
  if [ -n "$body" ]; then
    out=$(curl -s -w '\n%{http_code}' -X "$method" -H 'Content-Type: application/json' -d "$body" "$BASE$path")
  else
    out=$(curl -s -w '\n%{http_code}' -X "$method" "$BASE$path")
  fi
  code=$(tail -n1 <<<"$out"); resp=$(sed '$d' <<<"$out")
  echo "[$code] $method $path -> $resp"
  if [ "$code" != "$want" ]; then echo "FAIL: $name (want $want)"; exit 1; fi
  pass=$((pass+1))
}
check "health"            200 GET    /.well-known/health
check "list instructors"  200 GET    /instructors
check "book lesson"       201 POST   /bookings '{"guest_name":"Priya","instructor_id":1,"date":"2026-10-05","slot":"morning"}'
check "double booking"    409 POST   /bookings '{"guest_name":"Karan","instructor_id":1,"date":"2026-10-05","slot":"morning"}'
check "other slot ok"     201 POST   /bookings '{"guest_name":"Karan","instructor_id":1,"date":"2026-10-05","slot":"evening"}'
check "bad slot"          400 POST   /bookings '{"guest_name":"Anu","instructor_id":2,"date":"2026-10-05","slot":"midnight"}'
check "bad date"          400 POST   /bookings '{"guest_name":"Anu","instructor_id":2,"date":"05-10-2026","slot":"morning"}'
check "missing fields"    400 POST   /bookings '{"guest_name":"Anu"}'
check "unknown instructor" 404 POST  /bookings '{"guest_name":"Anu","instructor_id":99,"date":"2026-10-05","slot":"morning"}'
check "list by date"      200 GET    '/bookings?date=2026-10-05'
check "cancel"            204 DELETE /bookings/1
check "cancel again"      404 DELETE /bookings/1
echo "All $pass checks passed"

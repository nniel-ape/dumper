#!/usr/bin/env fish
# Decrypt and export secrets to environment
# Usage: source scripts/load-secrets.fish

set -l script_dir (dirname (status filename))
set -l project_root (dirname $script_dir)
set -l secrets_file "$project_root/secrets/prod.yaml"

if not test -f $secrets_file
    echo "Error: $secrets_file not found" >&2
    exit 1
end

for line in (sops -d --output-type dotenv $secrets_file 2>/dev/null)
    set -l parts (string split -m1 '=' $line)
    if test (count $parts) -eq 2
        set -l key (string upper $parts[1])
        set -l value $parts[2]
        set -gx $key $value
        echo "Exported: $key"
    end
end

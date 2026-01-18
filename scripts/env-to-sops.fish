#!/usr/bin/env fish
# Convert .env file to SOPS-encrypted YAML
# Usage: ./scripts/env-to-sops.fish [input.env] [output.yaml]

set -l input_file (test -n "$argv[1]" && echo $argv[1] || echo ".env")
set -l output_file (test -n "$argv[2]" && echo $argv[2] || echo "secrets/prod.yaml")

if not test -f $input_file
    echo "Error: $input_file not found" >&2
    exit 1
end

# Ensure secrets directory exists
mkdir -p (dirname $output_file)

# Convert .env to YAML (lowercase keys)
echo "Converting $input_file to $output_file..."
begin
    for line in (cat $input_file | grep -v '^#' | grep -v '^$')
        set -l parts (string split -m1 '=' $line)
        if test (count $parts) -eq 2
            set -l key (string lower $parts[1])
            set -l value $parts[2]
            # Remove surrounding quotes if present
            set value (string trim -c '"' $value)
            set value (string trim -c "'" $value)
            echo "$key: \"$value\""
        end
    end
end > $output_file

echo "Encrypting with SOPS..."
sops -e -i $output_file

echo "Done! Encrypted secrets saved to $output_file"

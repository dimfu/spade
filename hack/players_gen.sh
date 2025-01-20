#!/bin/bash

output_file="./database/seeds/players.sql"

generate_uuid() {
  local N B T
  for (( N=0; N < 16; ++N )); do
    B=$(( RANDOM%256 ))
    if (( N == 6 )); then
      printf '4%x' $(( B%16 ))
    elif (( N == 8 )); then
      T=$(( B%64 ))
      printf '%x%x' $(( T%4+8 )) $(( T%16 ))
    else
      printf '%02x' $B
    fi
    case $N in
      3|5|7|9) printf '-' ;;
    esac
  done
}

echo "START TRANSACTION;" > "$output_file"
echo "INSERT INTO players (id, name, discord_id) VALUES" >> "$output_file"

for i in {1..32}; do
  id=$(generate_uuid)
  name=$(fakedata -l 1 username)
  discord_id=$(generate_uuid)

  if [[ $i -lt 32 ]]; then
    echo "('$id', '$name', '$discord_id')," >> "$output_file"
  else
    echo "('$id', '$name', '$discord_id');" >> "$output_file"
  fi
done

echo "COMMIT;" >> "$output_file"

echo "SQL file '$output_file' has been generated."
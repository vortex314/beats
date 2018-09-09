set -x
rm -f data/registry
./filebeat -e -c fb.yml

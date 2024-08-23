#!/bin/bash

package_name="nexaai"
repo_name="nexa-sdk"

# Get output directory or default to index/whl/cpu
output_dir=${1:-"index/whl/cpu"}

# Create output directory
mkdir -p $output_dir

# Change to output directory
pushd $output_dir

# Create an index html file
echo "<!DOCTYPE html>" > index.html
echo "<html>" >> index.html
echo "  <head></head>" >> index.html
echo "  <body>" >> index.html
echo "    <a href=\"${package_name}/\">${package_name}</a>" >> index.html
echo "    <br>" >> index.html
echo "  </body>" >> index.html
echo "</html>" >> index.html
echo "" >> index.html

# Create ${package_name} directory
mkdir -p ${package_name}

# Change to ${package_name} directory
pushd ${package_name}

# Create an index html file
echo "<!DOCTYPE html>" > index.html
echo "<html>" >> index.html
echo "  <body>" >> index.html
echo "    <h1>Links for ${package_name}</h1>" >> index.html

# Get all releases
releases=$(curl -s https://api.github.com/repos/NexaAI/${repo_name}/releases | jq -r .[].tag_name)

# Get pattern from second arg or default to valid python package version pattern
pattern=${2:-"^[v]?[0-9]+\.[0-9]+\.[0-9]+$"}

# Filter releases by pattern
releases=$(echo $releases | tr ' ' '\n' | grep -E $pattern)

# For each release, get all assets
for release in $releases; do
    assets=$(curl -s https://api.github.com/repos/NexaAI/${repo_name}/releases/tags/$release | jq -r .assets)
    # Get release version from release ie v0.1.0-cu121 -> v0.1.0
    release_version=$(echo $release | grep -oE "^[v]?[0-9]+\.[0-9]+\.[0-9]+")
    echo "    <h2>$release_version</h2>" >> index.html
    for asset in $(echo $assets | jq -r .[].browser_download_url); do
        if [[ $asset == *".whl" ]]; then
            echo "    <a href=\"$asset\">$asset</a>" >> index.html
            echo "    <br>" >> index.html
        fi
    done
done

echo "  </body>" >> index.html
echo "</html>" >> index.html
echo "" >> index.html

# Extract the version from the output_dir (e.g., "cpu" from "index/whl/cpu")
version=$(basename "$output_dir")

# If this is the CPU version, create the root index.html
if [ "$version" == "cpu" ]; then
    # Return to original directory
    popd
    popd

    root_dir=$(dirname "$output_dir")
    echo "<!DOCTYPE html>" > "$root_dir/index.html"
    echo "<html>" >> "$root_dir/index.html"
    echo "  <head></head>" >> "$root_dir/index.html"
    echo "  <body>" >> "$root_dir/index.html"
    echo "    <h1>NEXAAI SDK Python Wheels</h1>" >> "$root_dir/index.html"
    echo "    <a href=\"cpu/\">CPU</a><br>" >> "$root_dir/index.html"
    echo "    <a href=\"metal/\">Metal</a><br>" >> "$root_dir/index.html"
    # echo "    <a href=\"cu121/\">CUDA 12.1</a><br>" >> "$root_dir/index.html"
    # echo "    <a href=\"cu122/\">CUDA 12.2</a><br>" >> "$root_dir/index.html"
    # echo "    <a href=\"cu123/\">CUDA 12.3</a><br>" >> "$root_dir/index.html"
    echo "    <a href=\"cu124/\">CUDA 12.4</a><br>" >> "$root_dir/index.html"
    echo "  </body>" >> "$root_dir/index.html"
    echo "</html>" >> "$root_dir/index.html"
fi
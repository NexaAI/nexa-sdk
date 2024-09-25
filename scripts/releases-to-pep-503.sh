#!/bin/bash

package_name="nexaai"
repo_name="nexa-sdk"

# Get output directory or default to index/whl/cpu
output_dir=${1:-"index/whl/cpu"}

# Create output directory
echo "Creating output directory: $output_dir"
mkdir -p $output_dir

# Change to output directory
echo "Changing to output directory: $output_dir"
pushd $output_dir

# Create an index html file
echo "Creating index.html in $output_dir"
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
echo "Creating package directory: ${package_name}"
mkdir -p ${package_name}

# Change to ${package_name} directory
echo "Changing to package directory: ${package_name}"
pushd ${package_name}

# Create an index html file
echo "Creating index.html in ${package_name} directory"
echo "<!DOCTYPE html>" > index.html
echo "<html>" >> index.html
echo "  <body>" >> index.html
echo "    <h1>Links for ${package_name}</h1>" >> index.html

# Get all releases
echo "Fetching all releases from GitHub for repository: ${repo_name}"
releases=$(curl -s https://api.github.com/repos/NexaAI/${repo_name}/releases | jq -r .[].tag_name)

# Output all retrieved releases for debugging
echo "All releases retrieved: $releases"

# Get pattern from second arg or default to valid python package version pattern
# pattern example 1 : v0.1.0
# pattern example 2 : 0.1.0
# pattern example 3 : 0.1.0-cu121
# pattern example 4 : v0.0.0.1
# pattern example 5 : v0.0.8.1
pattern=${2:-"^[v]?[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?(-[a-zA-Z0-9]+)?$"}

# Filter releases by pattern
echo "Filtering releases with pattern: $pattern"
releases=$(echo $releases | tr ' ' '\n' | grep -E $pattern)

# Output filtered releases for debugging
echo "Filtered releases: $releases"

# For each release, get all assets
for release in $releases; do
    echo "Processing release: $release"
    assets=$(curl -s https://api.github.com/repos/NexaAI/${repo_name}/releases/tags/$release | jq -r .assets)
    
    # Extract full release version without removing any segments
    release_version=$(echo $release | grep -oE "^[v]?[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?(-[a-zA-Z0-9]+)?$")
    
    # Debugging output for each release
    echo "Extracted version: $release_version"
    echo "Assets found for release $release_version:"
    echo $assets | jq -r .[].browser_download_url

    echo "    <h2>$release_version</h2>" >> index.html
    for asset in $(echo $assets | jq -r .[].browser_download_url); do
        if [[ $asset == *".whl" ]]; then
            echo "Adding wheel to index: $asset"
            echo "    <a href=\"$asset\">$asset</a>" >> index.html
            echo "    <br>" >> index.html
        else
            echo "Skipping non-wheel asset: $asset"
        fi
    done
done

echo "Closing HTML tags in index.html"
echo "  </body>" >> index.html
echo "</html>" >> index.html
echo "" >> index.html

# Extract the version from the output_dir (e.g., "cpu" from "index/whl/cpu")
version=$(basename "$output_dir")
echo "Extracted version from output directory: $version"

# If this is the CPU version, create the root index.html
if [ "$version" == "cpu" ]; then
    echo "Creating root index.html since version is CPU"
    
    # Return to original directory
    popd
    popd

    root_dir=$(dirname "$output_dir")
    echo "Root directory: $root_dir"
    echo "Creating index.html in root directory"
    
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
    echo "    <a href=\"rocm621/\">ROCm 6.2.1</a><br>" >> "$root_dir/index.html"
    echo "  </body>" >> "$root_dir/index.html"
    echo "</html>" >> "$root_dir/index.html"
fi

echo "Script execution completed."

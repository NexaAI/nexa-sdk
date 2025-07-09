require 'erb'

# --- Configuration ---
MANIFEST_DATA = ENV.fetch('MANIFEST_DATA')
RELEASE_TAG = ENV.fetch('RELEASE_TAG')
VERSION = RELEASE_TAG.delete_prefix('v')
RELEASE_REPO = ENV.fetch('RELEASE_REPOSITORY')
FORMULA_FILENAME = "Formula/nexa.rb"

# --- Mappings ---
OS_VERSION_MAP = {
  "macos-15" => { symbol: :sequoia, name: "on_sequoia" },
  "macos-14" => { symbol: :sonoma, name: "on_sonoma" },
  "macos-13" => { symbol: :ventura, name: "on_ventura" },
}.freeze

OS_ARCH_MAP = {
  "macos-15" => { symbol: :arm, name: "on_arm" },
  "macos-14" => { symbol: :arm, name: "on_arm" },
  "macos-13" => { symbol: :intel, name: "on_intel" },
}.freeze

# --- Data Parsing ---
puts "Parsing manifest data..."
assets = {}
MANIFEST_DATA.strip.split("\n").each do |line|
  # Manifest format: os_version;backend;package_name;sha256
  os_version, backend, package_name, sha256 = line.split(';')

  os_info = OS_VERSION_MAP[os_version]
  arch_info = OS_ARCH_MAP[os_version]
  next unless os_info && arch_info

  # Initialize nested hash if not present
  assets[os_info[:name]] ||= {}
  assets[os_info[:name]][arch_info[:name]] ||= {}

  # Store asset data
  assets[os_info[:name]][arch_info[:name]][backend] = {
    url: "https://github.com/#{RELEASE_REPO}/releases/download/#{RELEASE_TAG}/#{package_name}",
    sha256: sha256
  }
end

# Make assets available to the ERB template
@assets = assets
@version = VERSION
@repo_owner_and_name = RELEASE_REPO

# --- ERB Template ---
# This template precisely matches your desired formula structure.
TEMPLATE = <<~ERB
  # typed: false
  # frozen_string_literal: true

  class Nexa < Formula
    desc "A powerful CLI for the NexaAI ecosystem"
    homepage "https://github.com/<%= @repo_owner_and_name %>"
    version "<%= @version %>"
    license "MIT"

    option "with-mlx", "Install with the MLX backend instead of the default Llama-cpp-metal backend"

  <%- @assets.keys.sort.reverse.each do |os_block_name| -%>
  <%-   os_assets = @assets[os_block_name] -%>
    <%= os_block_name %> do
  <%-   os_assets.each do |arch_block_name, backend_assets| -%>
      <%= arch_block_name %> do
  <%-     if backend_assets['mlx'] -%>
        if build.with? "mlx"
          url "<%= backend_assets['mlx'][:url] %>"
          sha256 "<%= backend_assets['mlx'][:sha256] %>"
        else
          url "<%= backend_assets['llama-cpp-metal'][:url] %>"
          sha256 "<%= backend_assets['llama-cpp-metal'][:sha256] %>"
        end
  <%-     else -%>
        # This OS/Arch only supports the Llama-cpp-metal backend
        url "<%= backend_assets['llama-cpp-metal'][:url] %>"
        sha256 "<%= backend_assets['llama-cpp-metal'][:sha256] %>"
  <%-     end -%>
      end
  <%-   end -%>
    end
  <%- end -%>

  def install
      libexec.install "nexa"
      libexec.install "nexa-cli"
      libexec.install "lib"

      chmod "+x", libexec/"nexa"
      chmod "+x", libexec/"nexa-cli"

      (bin/"nexa").write <<~EOS
        #!/bin/bash
        export DYLD_LIBRARY_PATH="\#{libexec}/lib"
        exec "\#{libexec}/nexa" "$@"
      EOS

      (bin/"nexa-cli").write <<~EOS
        #!/bin/bash
        export DYLD_LIBRARY_PATH="\#{libexec}/lib"
        exec "\#{libexec}/nexa-cli" "$@"
      EOS
    end

    test do
      assert_match "version", shell_output("\#{bin}/nexa --version")
    end
  end
ERB

# --- File Generation ---
puts "Generating #{FORMULA_FILENAME}..."
renderer = ERB.new(TEMPLATE, trim_mode: "-")
output = renderer.result(binding)

File.write(FORMULA_FILENAME, output)
puts "Successfully wrote updated formula to #{FORMULA_FILENAME}"
#!/bin/bash

# 🎨 Terminal colors
readonly COLOR_RESET='\033[0m'
readonly COLOR_RED='\033[31m'
readonly COLOR_GREEN='\033[32m'
readonly COLOR_YELLOW='\033[33m'
readonly COLOR_BLUE='\033[34m'
readonly COLOR_MAGENTA='\033[35m'
readonly COLOR_CYAN='\033[36m'

# 🌐 Constants
readonly GO_VERSIONS_URL="https://go.dev/dl/?mode=json"
readonly GO_DOWNLOAD_BASE="https://go.dev/dl/"
readonly DEFAULT_INSTALL_DIR="/usr/local"
readonly ARCH_ARMV6L="armv6l"

# Helper function to print with colors
print_color() {
    local color=$1
    local message=$2
    printf "${color}%s${COLOR_RESET}" "$message"
    return 0
}

# Helper function to print with colors and newline
print_color_ln() {
    local color=$1
    local message=$2
    printf "${color}%s${COLOR_RESET}\n" "$message"
    return 0
}

# Helper function to get current Go version
get_current_go_version() {
    if command -v go >/dev/null 2>&1; then
        go version 2>/dev/null | awk '{print $3}'
    else
        echo ""
    fi
    return 0
}

# Helper function to detect current shell
detect_shell() {
    local shell_path="${SHELL:-unknown}"
    if [[ "$shell_path" == "unknown" ]]; then
        echo "unknown"
        return
    fi
    
    local shell_name=$(basename "$shell_path")
    case "$shell_name" in
        "zsh") echo "zsh 🚀" ;;
        "bash") echo "bash 🐚" ;;
        "fish") echo "fish 🐠" ;;
        *) echo "$shell_name" ;;
    esac
}

# Print banner
print_banner() {
    echo ""
    print_color "$COLOR_MAGENTA" "╔════════════════════════════════════════╗"
    echo ""
    printf "║   🚀 Go Version Manager (GVM) "
    print_color "$COLOR_YELLOW" "v1.2.0"
    print_color "$COLOR_MAGENTA" "   ║"
    echo ""
    print_color "$COLOR_MAGENTA" "╚════════════════════════════════════════╝"
    echo ""
    return 0
}

# Print system information
print_system_info() {
    echo ""
    print_color "$COLOR_CYAN" "🖥️  System Information"
    echo ""
    
    # OS
    local os_name=$(uname -s | tr '[:upper:]' '[:lower:]')
    printf "  OS:           "
    print_color "$COLOR_YELLOW" "$os_name"
    echo ""
    
    # Architecture
    local arch=$(uname -m)
    case "$arch" in
        "x86_64") arch="amd64" ;;
        "i386"|"i686") arch="386" ;;
        "armv6l"|"armv7l") arch="$ARCH_ARMV6L" ;;
        "aarch64") arch="arm64" ;;
        *) ;;
    esac
    printf "  Architecture: "
    print_color "$COLOR_YELLOW" "$arch"
    echo ""
    
    # CPUs
    local cpus=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "unknown")
    printf "  CPUs:         "
    print_color "$COLOR_YELLOW" "$cpus"
    echo ""
    
    # Current Go version
    local current_version=$(get_current_go_version)
    printf "  Go Version:   "
    if [[ -n "$current_version" ]]; then
        print_color "$COLOR_YELLOW" "$current_version"
    else
        print_color "$COLOR_RED" "Not installed"
    fi
    echo ""
    
    # Current shell
    local shell=$(detect_shell)
    printf "  Shell:        "
    print_color "$COLOR_YELLOW" "$shell"
    echo ""
    return 0
}

# Get latest stable Go version
get_latest_stable_version() {
    local versions_json
    if command -v curl >/dev/null 2>&1; then
        versions_json=$(curl -s --connect-timeout 15 --proto-default https --proto-redir =https "$GO_VERSIONS_URL" 2>/dev/null)
    elif command -v wget >/dev/null 2>&1; then
        versions_json=$(wget -q --max-redirect=0 --https-only --timeout=15 -O - "$GO_VERSIONS_URL" 2>/dev/null)
    else
        return 1
    fi
    
    if [[ -z "$versions_json" ]]; then
        return 1
    fi
    
    # Parse JSON to find first stable version using grep and cut
    local version
    version=$(echo "$versions_json" | grep '"version"' | head -1 | cut -d'"' -f4)
    
    if [[ -n "$version" ]]; then
        echo "$version"
        return 0
    else
        return 1
    fi
}

# Download with progress
download_with_progress() {
    local url=$1
    local output_file=$2
    
    if command -v curl >/dev/null 2>&1; then
        curl -L --proto-redir =https --progress-bar -o "$output_file" "$url"
    elif command -v wget >/dev/null 2>&1; then
        wget --https-only --progress=bar:force -O "$output_file" "$url"
    else
        print_color "$COLOR_RED" "\n❌ Error: Neither curl nor wget found"
        return 1
    fi
}

# Clean previous installation
clean_previous_install() {
    local install_dir=$1
    local go_path="$install_dir"
    
    if [[ -d "$go_path" ]]; then
        echo ""
        print_color "$COLOR_BLUE" "♻️  Removing previous installation..."
        echo ""
        rm -rf "$go_path"
    fi
    return 0
}

# Extract tar.gz
extract_tar_gz() {
    local src=$1
    local dst=$2
    
    if ! tar -xzf "$src" -C "$dst" --strip-components=1; then
        return 1
    fi
    return 0
}

# Update PATH in shell config
update_path() {
    local install_dir=$1
    local go_bin="$install_dir/bin"
    local go_root="$install_dir"
    
    local rc_file=""
    local path_addition=""
    
    # Determine shell config file
    case "$(detect_shell)" in
        "zsh"*) 
            rc_file="$HOME/.zshrc"
            path_addition="export GOROOT=$go_root\nexport PATH=$go_bin:\$PATH"
            ;;
        "bash"*) 
            rc_file="$HOME/.bashrc"
            path_addition="export GOROOT=$go_root\nexport PATH=$go_bin:\$PATH"
            ;;
        "fish"*) 
            rc_file="$HOME/.config/fish/config.fish"
            path_addition="set -gx GOROOT $go_root\nset -gx PATH $go_bin \$PATH"
            ;;
        *) 
            echo ""
            print_color "$COLOR_YELLOW" "⚠️  Unknown shell. Please manually add $go_bin to your PATH"
            echo ""
            return 1
            ;;
    esac
    
    # Check if already added
    if [[ -f "$rc_file" ]] && grep -q "$go_bin" "$rc_file"; then
        return 0
    fi
    
    # Create directory if needed (for fish)
    mkdir -p "$(dirname "$rc_file")"
    
    # Add to shell config
    echo -e "$path_addition" >> "$rc_file"
    echo ""
    print_color "$COLOR_GREEN" "➕ Added to PATH in $rc_file"
    echo ""
    
    return 0
}

# Verify installation
verify_installation() {
    local expected_version=$1
    echo ""
    print_color "$COLOR_BLUE" "🔍 Verifying installation..."
    echo ""
    
    # Try using go command from PATH
    if ! command -v go >/dev/null 2>&1; then
        print_color "$COLOR_RED" "❌ Go command not found in PATH"
        echo ""
        
        # Try direct path if user installation
        if [[ $(id -u) -ne 0 ]]; then
            local direct_go_path="$HOME/go/bin/go"
            if [[ -f "$direct_go_path" ]]; then
                print_color "$COLOR_YELLOW" "💡 Found Go at $direct_go_path but not in PATH"
                echo ""
            fi
        fi
        return
    fi
    
    local current_output=$(go version 2>/dev/null)
    if [[ "$current_output" != *"$expected_version"* ]]; then
        print_color "$COLOR_YELLOW" "⚠️  Version mismatch: expected $expected_version, got $current_output"
        echo ""
        return
    fi
    
    print_color "$COLOR_GREEN" "✅ $current_output"
    echo ""
}

# Install Go
install_go() {
    local version=$1
    
    # Determine system parameters
    local os_name=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    # Map architecture names
    case "$arch" in
        "x86_64") arch="amd64" ;;
        "i386"|"i686") arch="386" ;;
        "armv6l"|"armv7l") arch="$ARCH_ARMV6L" ;;
        "aarch64") arch="arm64" ;;
        *) ;;
    esac
    
    # Construct download URL
    local pkg_name="${version}.${os_name}-${arch}.tar.gz"
    local download_url="${GO_DOWNLOAD_BASE}${pkg_name}"
    local temp_file="/tmp/$pkg_name"
    
    echo ""
    print_color "$COLOR_BLUE" "📦 Downloading Go $version..."
    echo ""
    printf "  From: "
    print_color "$COLOR_YELLOW" "$download_url"
    echo ""
    
    # Download the package
    if ! download_with_progress "$download_url" "$temp_file"; then
        echo ""
        print_color "$COLOR_RED" "❌ Download failed"
        echo ""
        exit 1
    fi
    
    # Determine install location
    local install_dir="$DEFAULT_INSTALL_DIR"
    if [[ $(id -u) -ne 0 ]]; then
        install_dir="$HOME/go"
        echo ""
        print_color "$COLOR_YELLOW" "⚠️  Installing as non-root user to: "
        print_color "$COLOR_CYAN" "$install_dir"
        echo ""
    fi
    
    # Clean previous installation
    clean_previous_install "$install_dir"
    
    # Create install directory
    mkdir -p "$install_dir"
    
    # Extract the archive
    echo ""
    print_color "$COLOR_BLUE" "⚙️  Installing to $install_dir..."
    echo ""
    if ! extract_tar_gz "$temp_file" "$install_dir"; then
        print_color "$COLOR_RED" "❌ Extraction failed"
        echo ""
        rm -f "$temp_file"
        exit 1
    fi
    
    # Update PATH in shell config
    if ! update_path "$install_dir"; then
        print_color "$COLOR_YELLOW" "⚠️  Couldn't update PATH automatically"
        echo ""
        print_color "$COLOR_YELLOW" "  Please add $install_dir/bin to your PATH"
        echo ""
    fi
    
    # Clean up
    rm -f "$temp_file"
    return 0
}

# Print success message
print_success() {
    local version=$1
    local shell=$(detect_shell)
    local shell_name=$(echo "$shell" | awk '{print $1}')
    
    local config_file
    case "$shell_name" in
        "zsh") config_file="~/.zshrc" ;;
        "bash") config_file="~/.bashrc" ;;
        "fish") config_file="~/.config/fish/config.fish" ;;
        *) config_file="your shell config" ;;
    esac
    
    echo ""
    print_color "$COLOR_GREEN" "🎉 Successfully installed Go $version!"
    echo ""
    print_color "$COLOR_YELLOW" "To start using Go, you may need to:"
    echo ""
    printf "  - Restart your terminal\n"
    printf "  - Run "
    print_color "$COLOR_CYAN" "source $config_file"
    echo ""
    printf "  - Test with "
    print_color "$COLOR_CYAN" "go version"
    echo ""
    echo ""
    return 0
}

# Exit with error
exit_with_error() {
    local message=$1
    echo ""
    print_color "$COLOR_RED" "❌ Error: $message"
    echo ""
    exit 1
}

# Main function
main() {
    print_banner
    print_system_info
    
    # Get latest version
    echo ""
    print_color "$COLOR_BLUE" "🔍 Fetching latest Go version..."
    echo ""
    
    local version
    version=$(get_latest_stable_version)
    
    if [[ -z "$version" ]]; then
        print_color "$COLOR_RED" "❌ Error: Failed to fetch version information"
        echo ""
        exit 1
    fi
    
    print_color "$COLOR_GREEN" "✅ Found stable version: "
    print_color "$COLOR_YELLOW" "$version"
    echo ""
    
    # Check current Go version
    local current_version=$(get_current_go_version)
    if [[ "$current_version" == "$version" ]]; then
        echo ""
        print_color "$COLOR_GREEN" "✔️  Go $version is already installed. No update needed."
        echo ""
        exit 0
    fi
    
    # Install Go
    install_go "$version"
    
    # Verify installation
    verify_installation "$version"
    
    # Print success message
    print_success "$version"
}

# Run main function
main "$@"

#!/bin/bash

# AUR Package Update Script for hpaper with dependency management and Go build
# Usage: ./update_aur.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PACKAGE_NAME="hpaper"
GITHUB_USER="Hcode00"
AUR_REPO="ssh://aur@aur.archlinux.org/${PACKAGE_NAME}.git"

echo -e "${BLUE}=== AUR Package Update Script for ${PACKAGE_NAME} ===${NC}"

# Ask what type of update
echo -e "${YELLOW}What type of update? ${NC}"
echo "1) Version update only"
echo "2) Dependencies update only" 
echo "3) Both version and dependencies"
echo "4) Fix build function only"
read -r UPDATE_TYPE

case $UPDATE_TYPE in
    1|3)
        echo -e "${YELLOW}Enter the new version (e.g., 0.5, 0.5.0, 1.0):${NC}"
        read -r NEW_VERSION
        
        if [[ -z "$NEW_VERSION" ]]; then
            echo -e "${RED}Error: Version cannot be empty${NC}"
            exit 1
        fi
        
        # Check GitHub release
        echo -e "${BLUE}Checking if GitHub release v${NEW_VERSION} exists...${NC}"
        RELEASE_URL="https://api.github.com/repos/${GITHUB_USER}/${PACKAGE_NAME}/releases/tags/v${NEW_VERSION}"
        HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$RELEASE_URL")
        
        if [[ "$HTTP_STATUS" != "200" ]]; then
            echo -e "${RED}Error: GitHub release v${NEW_VERSION} not found!${NC}"
            exit 1
        fi
        echo -e "${GREEN}GitHub release v${NEW_VERSION} found!${NC}"
        ;;
esac

# Clone AUR repo
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

echo -e "${BLUE}Cloning AUR repository...${NC}"
if ! GIT_SSH_COMMAND="ssh -o StrictHostKeyChecking=accept-new" git clone "$AUR_REPO"; then
    echo -e "${RED}Error: Failed to clone AUR repository${NC}"
    exit 1
fi

cd "$PACKAGE_NAME"

# Show current state
CURRENT_VERSION=$(grep "^pkgver=" PKGBUILD | cut -d'=' -f2)
CURRENT_PKGREL=$(grep "^pkgrel=" PKGBUILD | cut -d'=' -f2)
echo -e "${BLUE}Current: ${CURRENT_VERSION}-${CURRENT_PKGREL}${NC}"

# Backup
cp PKGBUILD PKGBUILD.backup

# Update version if requested
if [[ $UPDATE_TYPE == "1" || $UPDATE_TYPE == "3" ]]; then
    echo -e "${BLUE}Updating version to ${NEW_VERSION}...${NC}"
    sed -i "s/^pkgver=.*/pkgver=${NEW_VERSION}/" PKGBUILD
    sed -i "s/^pkgrel=.*/pkgrel=1/" PKGBUILD
    sed -i "s|archive/v.*\.tar\.gz|archive/v${NEW_VERSION}.tar.gz|g" PKGBUILD
    
    # Update checksums
    echo -e "${BLUE}Downloading source and generating checksums...${NC}"
    rm -f *.tar.gz*
    
    if ! makepkg --geninteg > checksums.tmp 2>/dev/null; then
        echo -e "${RED}Error: Could not download source or generate checksums${NC}"
        exit 1
    fi
    
    NEW_CHECKSUMS=$(cat checksums.tmp)
    sed -i '/^[[:space:]]*\(md5\|sha[0-9]*\|b2\)sums=/,/^[[:space:]]*)/d' PKGBUILD
    echo "$NEW_CHECKSUMS" >> PKGBUILD
    echo -e "${GREEN}Checksums updated successfully${NC}"
fi

# Update dependencies if requested
if [[ $UPDATE_TYPE == "2" || $UPDATE_TYPE == "3" || $UPDATE_TYPE == "4" ]]; then
    echo -e "${BLUE}Updating dependencies and build function...${NC}"
    
    # Ensure Go is in makedepends
    if ! grep -q "makedepends.*go" PKGBUILD; then
        if grep -q "^makedepends=" PKGBUILD; then
            sed -i "s/^makedepends=(/makedepends=('go' /" PKGBUILD
        else
            sed -i "/^license=/a makedepends=('go')" PKGBUILD
        fi
        echo -e "${GREEN}Added 'go' to makedepends${NC}"
    fi
    
    # Check if optdepends already exists
    if grep -q "^optdepends=" PKGBUILD; then
        echo -e "${YELLOW}Optional dependencies already exist. Replace? (y/n):${NC}"
        read -r replace_deps
        if [[ "$replace_deps" == "y" || "$replace_deps" == "Y" ]]; then
            sed -i '/^optdepends=/,/^[^[:space:]]/{ /^optdepends=/d; /^[[:space:]]/d; }' PKGBUILD
        fi
    else
        replace_deps="y"
    fi
    
    # Add new optdepends
    if [[ "$replace_deps" == "y" || "$replace_deps" == "Y" ]]; then
        sed -i "/^makedepends=/a optdepends=('swaybg: Backend for wlroots-based compositors (Sway, river, etc.)'\\
           'hyprpaper: Backend for Hyprland window manager'\\
           'plasma-workspace: Provides Plasma wallpaper tooling (KDE backend fallback)')" PKGBUILD
        echo -e "${GREEN}Optional dependencies added${NC}"
    fi
    
    # If only dependency/build update, increment pkgrel
    if [[ $UPDATE_TYPE == "2" || $UPDATE_TYPE == "4" ]]; then
        NEW_PKGREL=$((CURRENT_PKGREL + 1))
        sed -i "s/^pkgrel=.*/pkgrel=${NEW_PKGREL}/" PKGBUILD
        echo -e "${BLUE}Incremented pkgrel to ${NEW_PKGREL}${NC}"
    fi
fi

# Add/update build function
if [[ $UPDATE_TYPE == "1" || $UPDATE_TYPE == "2" || $UPDATE_TYPE == "3" || $UPDATE_TYPE == "4" ]]; then
    echo -e "${BLUE}Adding/updating build and package functions...${NC}"

    # To avoid leaving a broken PKGBUILD behind (missing build/package) on failures,
    # update functions atomically via a temp file.
    TMP_PKG=$(mktemp)
    cp PKGBUILD "$TMP_PKG"

    # Remove existing build, package, and prepare functions from the temp copy
    sed -i '/^prepare()/,/^}$/d' "$TMP_PKG"
    sed -i '/^build()/,/^}$/d' "$TMP_PKG"
    sed -i '/^package()/,/^}$/d' "$TMP_PKG"

    # Append build and package functions (prepare() is no longer needed; the GIF should not be shipped in release tarballs)
    cat >> "$TMP_PKG" << 'EOF'

build() {
  cd "${srcdir}/${pkgname}-${pkgver}"

  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  export CGO_LDFLAGS="${LDFLAGS}"
  export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw"

  go build -o hpaper .
}

package() {
  cd "${srcdir}/${pkgname}-${pkgver}"
  install -Dm755 hpaper "${pkgdir}/usr/bin/hpaper"
}
EOF

    if ! grep -q "^build()" "$TMP_PKG" || ! grep -q "^package()" "$TMP_PKG"; then
        echo -e "${RED}Error: Failed to generate build()/package() blocks${NC}"
        exit 1
    fi

    mv "$TMP_PKG" PKGBUILD
    echo -e "${GREEN}Build and package functions updated${NC}"
fi

# Debug: Let's check what the archive actually contains
VERSION_FOR_TEST=${NEW_VERSION:-$CURRENT_VERSION}
echo -e "${BLUE}Testing archive extraction for v${VERSION_FOR_TEST}...${NC}"
TEMP_TEST_DIR=$(mktemp -d)
cd "$TEMP_TEST_DIR"

# Download and extract to see the actual structure
curl -L -s "https://github.com/${GITHUB_USER}/${PACKAGE_NAME}/archive/v${VERSION_FOR_TEST}.tar.gz" -o test.tar.gz
tar -tf test.tar.gz | head -5
echo -e "${BLUE}Archive contains directories starting with:${NC}"
tar -tf test.tar.gz | head -1 | cut -d'/' -f1

cd "$TEMP_DIR/$PACKAGE_NAME"
rm -rf "$TEMP_TEST_DIR"

# Test build
echo -e "${BLUE}Testing package build...${NC}"
if ! makepkg --syncdeps --noconfirm --clean; then
    echo -e "${RED}Error: Package build failed!${NC}"
    echo -e "${YELLOW}Let's check what's actually in the extracted source:${NC}"
    
    # Extract source manually to debug
    makepkg --nobuild --syncdeps --noconfirm 2>/dev/null || true
    echo -e "${YELLOW}Contents of src directory:${NC}"
    ls -la src/ 2>/dev/null || echo "No src directory found"
    if [ -d src ]; then
        echo -e "${YELLOW}Subdirectories in src:${NC}"
        find src -maxdepth 2 -type d
    fi
    
    echo -e "${YELLOW}Check the PKGBUILD manually. Files are in: ${TEMP_DIR}/${PACKAGE_NAME}${NC}"
    exit 1
fi

echo -e "${GREEN}Build test successful!${NC}"

# Update .SRCINFO
echo -e "${BLUE}Updating .SRCINFO...${NC}"
makepkg --printsrcinfo > .SRCINFO

# Show changes
echo -e "${YELLOW}=== Changes to be committed ===${NC}"
git diff --no-index --no-prefix PKGBUILD.backup PKGBUILD || true

# Generate commit message
case $UPDATE_TYPE in
    1) DEFAULT_MSG="Update to version ${NEW_VERSION}" ;;
    2) DEFAULT_MSG="Add Go build function and optional dependencies" ;;
    3) DEFAULT_MSG="Update to version ${NEW_VERSION}, add build function and dependencies" ;;
    4) DEFAULT_MSG="Fix build function and dependencies" ;;
esac

echo -e "${YELLOW}Enter commit message (or press Enter for default):${NC}"
echo -e "${BLUE}Default: ${DEFAULT_MSG}${NC}"
read -r COMMIT_MSG

if [[ -z "$COMMIT_MSG" ]]; then
    COMMIT_MSG="$DEFAULT_MSG"
fi

# Final confirmation
echo -e "${YELLOW}Ready to commit and push? (y/n):${NC}"
read -r confirm

if [[ "$confirm" == "y" || "$confirm" == "Y" ]]; then
    echo -e "${BLUE}Committing changes...${NC}"
    git add PKGBUILD .SRCINFO
    git commit -m "$COMMIT_MSG"
    
    echo -e "${BLUE}Pushing to AUR...${NC}"
    git push
    
    echo -e "${GREEN}=== SUCCESS ===${NC}"
    echo -e "${GREEN}Package updated successfully!${NC}"
else
    echo -e "${YELLOW}Aborted. Changes are in: ${TEMP_DIR}/${PACKAGE_NAME}${NC}"
fi

# Cleanup
cd /
rm -rf "$TEMP_DIR"

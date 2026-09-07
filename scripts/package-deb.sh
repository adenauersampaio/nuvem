#!/bin/bash
#
# scripts/package-deb.sh
# Cria o pacote Debian (.deb) para o Nuvem.
#
# Uso:
#   ./scripts/package-deb.sh [versao]
#   Exemplo: ./scripts/package-deb.sh 0.1.0
#

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

RAW_VERSION="${1:-0.1.1}"
# Remove prefixo 'v' se fornecido (ex: v0.1.0 -> 0.1.0)
VERSION="${RAW_VERSION#v}"
# Converte -beta.1 para ~beta1 para conformidade estrita com o Debian versioning
DEB_VERSION="$(echo "$VERSION" | sed 's/-beta\./~beta/')"

ARCH="amd64"
PACKAGE_NAME="nuvem"
BUILD_DIR="build/deb/${PACKAGE_NAME}_${DEB_VERSION}_${ARCH}"
DIST_DIR="dist"

echo "==> Empacotando ${PACKAGE_NAME} versão ${DEB_VERSION} (${ARCH})..."

# 1. Limpeza prévia
rm -rf "build/deb"
mkdir -p "${BUILD_DIR}/DEBIAN"
mkdir -p "${BUILD_DIR}/usr/bin"
mkdir -p "${BUILD_DIR}/usr/share/applications"
mkdir -p "${BUILD_DIR}/usr/lib/systemd/user"
mkdir -p "${DIST_DIR}"
chmod 755 "${BUILD_DIR}/DEBIAN"


# 2. Compilação dos binários com otimizações de tamanho (-s -w)
echo "==> Compilando binários Go..."
CGO_ENABLED=0 go build -ldflags "-s -w" -o "${BUILD_DIR}/usr/bin/nuvem" ./cmd/nuvem
CGO_ENABLED=1 go build -ldflags "-s -w" -o "${BUILD_DIR}/usr/bin/nuvem-desktop" ./cmd/nuvem-desktop
chmod 755 "${BUILD_DIR}/usr/bin/nuvem" "${BUILD_DIR}/usr/bin/nuvem-desktop"

# 3. Instalação dos ícones
echo "==> Copiando ícones da aplicação..."
ICONS_SRC="internal/desktop/assets/icons"
ICON_SIZES=(16 32 48 64 120 128 256 512 1024)

for SIZE in "${ICON_SIZES[@]}"; do
  ICON_DEST_DIR="${BUILD_DIR}/usr/share/icons/hicolor/${SIZE}x${SIZE}/apps"
  mkdir -p "$ICON_DEST_DIR"
  if [ -f "${ICONS_SRC}/nuvem-${SIZE}.png" ]; then
    cp "${ICONS_SRC}/nuvem-${SIZE}.png" "${ICON_DEST_DIR}/io.github.adenauersampaio.Nuvem.png"
    chmod 644 "${ICON_DEST_DIR}/io.github.adenauersampaio.Nuvem.png"
  fi
done

# 4. Criação do lançador .desktop
echo "==> Gerando arquivo .desktop..."
cat << 'EOF' > "${BUILD_DIR}/usr/share/applications/io.github.adenauersampaio.Nuvem.desktop"
[Desktop Entry]
Type=Application
Name=Nuvem
Name[pt_BR]=Nuvem
Comment=Continuous cloud folder synchronization
Comment[pt_BR]=Sincronização contínua de pastas na nuvem
Exec=/usr/bin/nuvem-desktop
Icon=io.github.adenauersampaio.Nuvem
Terminal=false
Categories=Utility;FileTransfer;Network;
StartupNotify=true
EOF
chmod 644 "${BUILD_DIR}/usr/share/applications/io.github.adenauersampaio.Nuvem.desktop"

# 5. Criação da unidade de serviço systemd de usuário
echo "==> Gerando serviço systemd de usuário..."
cat << 'EOF' > "${BUILD_DIR}/usr/lib/systemd/user/nuvem.service"
[Unit]
Description=Nuvem continuous synchronization
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/bin/nuvem daemon
Restart=on-failure
RestartSec=15s

[Install]
WantedBy=default.target
EOF
chmod 644 "${BUILD_DIR}/usr/lib/systemd/user/nuvem.service"

# 6. Criação do arquivo DEBIAN/control
echo "==> Gerando metadados do pacote Debian..."
INSTALLED_SIZE=$(du -sk "${BUILD_DIR}/usr" | cut -f1)

cat << EOF > "${BUILD_DIR}/DEBIAN/control"
Package: ${PACKAGE_NAME}
Version: ${DEB_VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Depends: libc6, libgtk-4-1 (>= 4.6)
Installed-Size: ${INSTALLED_SIZE}
Maintainer: Adenauer Sampaio <adenauersampaio@gmail.com>
Homepage: https://adenauersampaio.github.io/nuvem/
Description: Continuous cloud folder synchronization for Linux
 Nuvem provides continuous, understandable cloud-folder synchronization on Linux.
 It offers a native GTK4 visual interface and a background daemon managed
 through systemd to synchronize local folders directly with Google Drive.
EOF
chmod 644 "${BUILD_DIR}/DEBIAN/control"

# 7. Scripts de gatilho (postinst e postrm)
cat << 'EOF' > "${BUILD_DIR}/DEBIAN/postinst"
#!/bin/sh
set -e

if [ "$1" = "configure" ]; then
    if which gtk-update-icon-cache >/dev/null 2>&1; then
        gtk-update-icon-cache -q -t -f /usr/share/icons/hicolor || true
    fi
    if which update-desktop-database >/dev/null 2>&1; then
        update-desktop-database -q /usr/share/applications || true
    fi
fi

exit 0
EOF
chmod 755 "${BUILD_DIR}/DEBIAN/postinst"

cat << 'EOF' > "${BUILD_DIR}/DEBIAN/postrm"
#!/bin/sh
set -e

if [ "$1" = "remove" ] || [ "$1" = "purge" ]; then
    if which gtk-update-icon-cache >/dev/null 2>&1; then
        gtk-update-icon-cache -q -t -f /usr/share/icons/hicolor || true
    fi
    if which update-desktop-database >/dev/null 2>&1; then
        update-desktop-database -q /usr/share/applications || true
    fi
fi

exit 0
EOF
chmod 755 "${BUILD_DIR}/DEBIAN/postrm"

# 8. Construção do arquivo .deb
find "${BUILD_DIR}" -type d -exec chmod 755 {} +
DEB_FILE="${DIST_DIR}/${PACKAGE_NAME}_${DEB_VERSION}_${ARCH}.deb"
echo "==> Gerando arquivo .deb final..."
dpkg-deb --build --root-owner-group "${BUILD_DIR}" "${DEB_FILE}"

echo "==> Sucesso! Pacote gerado em: ${DEB_FILE}"
ls -lh "${DEB_FILE}"

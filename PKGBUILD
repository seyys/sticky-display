pkgname='sticky-display'
_pkgname=${pkgname%-git}
pkgver=0.1
pkgrel=1
pkgdesc='Make all windows on a display sticky. For Xfce and other EWMH Compliant Window Managers.'
arch=('x86_64')
url='https://github.com/seyys/sticky-display'
license=('MIT')
depends=()
makedepends=()
optdepends=('xorg-server: with EWMH Complaint Window Managers')
provides=('sticky-display')
conflicts=()
source=("${_pkgname}::git+$url.git")
md5sums=('SKIP')

build() {
  cd "$_pkgname"
  export GOPATH="$srcdir"
  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  export CGO_LDFLAGS="${LDFLAGS}"
  export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw"
  go build
}

package() {
  cd "$_pkgname"
  install -Dm755 ./$_pkgname  "$pkgdir/usr/bin/$pkgname"
  install -Dm644 ./README.md  "$pkgdir/usr/share/doc/$_pkgname"
  install -Dm644 ./LICENSE    "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}

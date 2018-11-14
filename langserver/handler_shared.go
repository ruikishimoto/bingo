package langserver

import (
	"fmt"
	"strings"
	"sync"

	"github.com/saibing/bingo/langserver/internal/caches"
	"github.com/sourcegraph/jsonrpc2"
	"golang.org/x/tools/go/packages"

	"github.com/sourcegraph/ctxvfs"
)

// HandlerShared contains data structures that a build server and its
// wrapped lang server may share in memory.
type HandlerShared struct {
	Mu     sync.Mutex // guards all fields
	Shared bool       // true if this struct is shared with a build server
	FS     *AtomicFS  // full filesystem (mounts both deps and overlay)

	overlay *overlay // files to overlay
}



func (h *HandlerShared) getFindPackageFunc() caches.FindPackageFunc {
	return defaultFindPackageFunc
}

func defaultFindPackageFunc(packageCache *caches.PackageCache, importPath string) (*packages.Package, error) {
	if strings.HasPrefix(importPath, "/") {
		return nil, fmt.Errorf("import %q: cannot import absolute path", importPath)
	}
	return packageCache.Lookup(importPath), nil
}

func (h *HandlerShared) Reset(conn *jsonrpc2.Conn, useOSFS bool) error {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	h.overlay = newOverlay(conn)
	h.FS = NewAtomicFS()

	if useOSFS {
		// The overlay FS takes precedence, but we fall back to the OS
		// file system.
		h.FS.Bind("/", ctxvfs.OS("/"), "/", ctxvfs.BindAfter)
	}
	//h.FS.Bind("/", h.overlay.FS(), "/", ctxvfs.BindBefore)
	return nil
}

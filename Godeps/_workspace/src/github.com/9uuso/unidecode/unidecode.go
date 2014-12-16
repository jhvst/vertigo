// Package unidecode implements a unicode transliterator
// which replaces non-ASCII characters with their ASCII
// approximations.
package unidecode

import (
	"compress/zlib"
	"encoding/binary"
	"io"
	"strings"
	"sync"
	"unicode"

	"gopkgs.com/pool.v1"
)

const pooledCapacity = 64

var (
	mutex            sync.Mutex
	transliterations [65536][]rune

	slicePool  = pool.New(0)
	decoded    = false
	transCount = rune(len(transliterations))
	getUint16  = binary.LittleEndian.Uint16

package global

import (
	"context"
	"math/rand"
	"sync"

	cry "github.com/zicops/zicops-course-query/lib/crypto"
)

// some global variables commonly used
var (
	CTX             context.Context
	CryptSession    *cry.Cryptography
	Cancel          context.CancelFunc
	WaitGroupServer sync.WaitGroup
	Rand            *rand.Rand
)

// initializes global package to read environment variables as needed
func init() {
}

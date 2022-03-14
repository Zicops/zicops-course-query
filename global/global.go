package global

import (
	"context"
	"sync"

	"github.com/zicops/zicops-course-query/lib/db/cassandra"
	cry "github.com/zicops/zicops-course-query/lib/crypto"

)

// some global variables commonly used
var (
	CTX             context.Context
	CassSession     *cassandra.Cassandra
	CryptSession	*cry.Cryptography
	Cancel          context.CancelFunc
	WaitGroupServer sync.WaitGroup
)

// initializes global package to read environment variables as needed
func init() {
}

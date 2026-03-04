package balancer

// import "time"

type leastConn struct{
	pool *BackendPool
}

func NewLeastCount(pool *BackendPool)*leastConn{
	return &leastConn{
		pool: pool,
	}
}

func (lc *leastConn) NextBackend() *Backend{

	healthy:= lc.pool.GetHealthyBackends()

	if len(healthy)==0 {
		return nil
	}

	var selected *Backend
	minConn :=int64(-1)
	for _,b :=range healthy{
		conn :=b.ActiveConnections

		if minConn ==-1 || conn < minConn {
			minConn = conn
			selected=b
		}
	}
	return selected
}

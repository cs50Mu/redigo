package main

// Pipeline represents a redis pipeline
type Pipeline struct {
	conn      Conn
	pool      *ConnPool
	cmdBuffer [][]string
	cmdCnt    int
}

// AddCommand add redis command
func (p *Pipeline) AddCommand(command string, args ...string) {
	commands := append([]string{command}, args...)
	p.cmdBuffer = append(p.cmdBuffer, commands)
	p.cmdCnt++
}

// Exec a redis pipeline
// returns results of queued commands
func (p *Pipeline) Exec() ([]*Reply, error) {
	err := p.conn.SendBulkCommand(p.cmdBuffer)
	if err != nil {
		return nil, err
	}
	var res []*Reply
	for i := p.cmdCnt; i > 0; i-- {
		reply, err := p.conn.ReadResp()
		if err != nil {
			return nil, err
		}
		res = append(res, reply)
	}
	return res, nil
}

// Close a redis pipeline
func (p *Pipeline) Close() {
	p.pool.ReleaseConn(p.conn)
}

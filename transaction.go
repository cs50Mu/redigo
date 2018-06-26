package main

// Transaction represents a redis transaction
type Transaction struct {
	conn    Conn
	pool    *ConnPool
	started bool
}

// AddCommand adds one command to current transaction
func (tx *Transaction) AddCommand(command string, args ...string) error {
	commandSlice := append([]string{command}, args...)
	if !tx.started {
		tx.started = true
		if err := tx.conn.SendCommand("MULTI"); err != nil {
			return err
		}
		tx.conn.ReadResp()
	}
	tx.conn.SendCommand(commandSlice...)
	_, err := tx.conn.ReadResp()
	return err
}

// Watch marks the given keys to be watched for conditional execution of a transaction
func (tx *Transaction) Watch(keys ...string) error {
	commandSlice := append([]string{"WATCH"}, keys...)
	tx.conn.SendCommand(commandSlice...)
	_, err := tx.conn.ReadResp()
	return err
}

// Unwatch flushes all the previously watched keys for a transaction
func (tx *Transaction) Unwatch() error {
	tx.conn.SendCommand("UNWATCH")
	_, err := tx.conn.ReadResp()
	return err
}

// Exec executes all previously queued commands in a transaction and restores the connection state to normal.
func (tx *Transaction) Exec() ([]*Reply, error) {
	if err := tx.conn.SendCommand("EXEC"); err != nil {
		return nil, err
	}
	reply, err := tx.conn.ReadResp()
	if err != nil {
		return nil, err
	}
	return reply.arrayVal, nil
}

// Discard flushes all previously queued commands in a transaction and restores the connection state to normal.
func (tx *Transaction) Discard() error {
	tx.conn.SendCommand("DISCARD")
	_, err := tx.conn.ReadResp()
	return err
}

// Close the underlying connection of the transaction
func (tx *Transaction) Close() {
	tx.pool.ReleaseConn(tx.conn)
}

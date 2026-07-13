package ws

import "sync"

const outboundQueueCapacity = 64

type outboundEnvelope struct {
	payload    any
	closeAfter bool
}

type orderedSender struct {
	connection Connection
	onFailure  func(Connection)
	mu         sync.Mutex
	cond       *sync.Cond
	queue      []outboundEnvelope
	closed     bool
}

func newOrderedSender(connection Connection, onFailure func(Connection)) *orderedSender {
	sender := &orderedSender{connection: connection, onFailure: onFailure}
	sender.cond = sync.NewCond(&sender.mu)
	go sender.run()
	return sender
}

func (s *orderedSender) enqueue(envelope outboundEnvelope) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	if len(s.queue) >= outboundQueueCapacity {
		return false
	}
	s.queue = append(s.queue, envelope)
	s.cond.Signal()
	return true
}

func (s *orderedSender) stop() {
	s.mu.Lock()
	s.closed = true
	s.queue = nil
	s.cond.Broadcast()
	s.mu.Unlock()
}

func (s *orderedSender) run() {
	for {
		s.mu.Lock()
		for len(s.queue) == 0 && !s.closed {
			s.cond.Wait()
		}
		if s.closed {
			s.mu.Unlock()
			return
		}
		envelope := s.queue[0]
		s.queue = s.queue[1:]
		s.mu.Unlock()

		if envelope.payload != nil {
			if err := s.connection.SendJSON(envelope.payload); err != nil {
				s.onFailure(s.connection)
				return
			}
		}
		if envelope.closeAfter {
			s.onFailure(s.connection)
			return
		}
	}
}

type outboundDispatcher struct {
	mu        sync.Mutex
	senders   map[Connection]*orderedSender
	onFailure func(Connection)
}

func newOutboundDispatcher(onFailure func(Connection)) *outboundDispatcher {
	return &outboundDispatcher{senders: make(map[Connection]*orderedSender), onFailure: onFailure}
}

func (d *outboundDispatcher) send(connection Connection, payload any) bool {
	return d.enqueue(connection, outboundEnvelope{payload: payload})
}

func (d *outboundDispatcher) sendAndClose(connection Connection, payload any) bool {
	return d.enqueue(connection, outboundEnvelope{payload: payload, closeAfter: true})
}

func (d *outboundDispatcher) closeWhenDrained(connection Connection) bool {
	return d.enqueue(connection, outboundEnvelope{closeAfter: true})
}

func (d *outboundDispatcher) enqueue(connection Connection, envelope outboundEnvelope) bool {
	d.mu.Lock()
	sender := d.senders[connection]
	if sender == nil {
		sender = newOrderedSender(connection, d.fail)
		d.senders[connection] = sender
	}
	d.mu.Unlock()
	accepted := sender.enqueue(envelope)
	if !accepted {
		go d.fail(connection)
	}
	return accepted
}

func (d *outboundDispatcher) remove(connection Connection) {
	d.mu.Lock()
	sender := d.senders[connection]
	delete(d.senders, connection)
	d.mu.Unlock()
	if sender != nil {
		sender.stop()
	}
}

func (d *outboundDispatcher) fail(connection Connection) {
	d.remove(connection)
	_ = connection.Close()
	if d.onFailure != nil {
		d.onFailure(connection)
	}
}

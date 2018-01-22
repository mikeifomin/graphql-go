package ws

import (
  "encoding/json"
	"github.com/gorilla/websocket"
)

type ProtoObserver struct {
	conn *websocket.Conn
}

func (p *ProtoObserver) Next(id string, data json.RawMessage) {
	err := p.conn.WriteJSON(Message{
		Type:    GQL_DATA,
		Id:      id,
		Payload: data,
	})
	if err != nil {
		panic(err)
	}
}

// Error function run if graphql validation error
// see https://github.com/apollographql/subscriptions-transport-ws/blob/v0.9.5/PROTOCOL.md#gql_error
func (p *ProtoObserver) Error(id string, err error) {
	errW := p.conn.WriteJSON(Message{
		Type:    GQL_ERROR,
		Id:      id,
		Payload: json.RawMessage(err.Error()),
	})
	if errW != nil {
		panic(errW)
	}
}
func (p *ProtoObserver) Complete(id string) {
	err := p.conn.WriteJSON(Message{
		Type:    GQL_COMPLETE,
		Id:      id,
	})
	if err != nil {
		panic(err)
	}
}


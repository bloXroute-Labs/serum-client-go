package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/jsonrpc2"
)

func UnaryWSRequest[T any](conn *websocket.Conn, request string) (*T, error) {
	err := sendWSRequest(conn, request)
	if err != nil {
		return nil, err
	}

	return recvWSResult[T](conn)
}

func UnaryWSStream[T any](ctx context.Context, conn *websocket.Conn, request string, responseChan chan *T) error {
	err := sendWSRequest(conn, request)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				response, err := recvWSResult[T](conn)
				if err != nil {
					logrus.Error(err)
					continue
				}

				responseChan <- response
			}
		}
	}()

	return nil
}

func sendWSRequest(conn *websocket.Conn, request string) error {
	if err := conn.WriteMessage(websocket.TextMessage, []byte(request)); err != nil {
		return fmt.Errorf("error with sending message - %v", err)
	}
	return nil
}

func recvWSResult[T any](conn *websocket.Conn) (*T, error) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("error reading WS response - %v", err)
	}

	// extract the result
	var resp jsonrpc2.Response
	err = resp.UnmarshalJSON(msg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON response - %v", err)
	}
	resultBytes, err := resp.Result.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON result - %v", err)
	}

	var result T
	if err = json.Unmarshal(resultBytes, &result); err != nil {
		return nil, fmt.Errorf("error with unmarshalling message of type %T - %v", result, err)
	}
	return &result, nil
}
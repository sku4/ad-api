package mocks

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

func TestClient_SendMessage(t *testing.T) {
	type mockBehavior func(r *MockBotClient, message string, chatId int64)

	tests := []struct {
		name    string
		message string
		chatID  int64
		mockBehavior
		wantErr bool
	}{
		{
			name: "Send correct message",
			mockBehavior: func(r *MockBotClient, message string, chatId int64) {
				r.EXPECT().SendMessage(message, chatId).Return(nil)
			},
			message: "test 1",
			chatID:  15,
			wantErr: false,
		},
		{
			name: "Send incorrect message",
			mockBehavior: func(r *MockBotClient, message string, chatId int64) {
				r.EXPECT().SendMessage(message, chatId).Return(
					errors.New("message is empty"))
			},
			message: "",
			chatID:  15,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			tgClient := NewMockBotClient(c)
			tt.mockBehavior(tgClient, tt.message, tt.chatID)

			if err := tgClient.SendMessage(tt.message, tt.chatID); (err != nil) != tt.wantErr {
				t.Errorf("SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

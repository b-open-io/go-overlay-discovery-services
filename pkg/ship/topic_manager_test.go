package ship

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test helper functions

func createTestSHIPTopicManager() (*SHIPTopicManager, *MockSHIPStorageInterface) {
	mockStorage := new(MockSHIPStorageInterface)

	topicManager := NewSHIPTopicManager(mockStorage, nil)

	return topicManager, mockStorage
}

func createTestSHIPTopicManagerWithLookupService() (*SHIPTopicManager, *MockSHIPStorageInterface, *SHIPLookupService) {
	mockStorage := new(MockSHIPStorageInterface)

	lookupService := NewSHIPLookupService(mockStorage)
	topicManager := NewSHIPTopicManager(mockStorage, lookupService)

	return topicManager, mockStorage, lookupService
}

func createTestTopicMessage(topic, messageID string, payload interface{}) TopicMessage {
	return TopicMessage{
		Topic:      topic,
		Payload:    payload,
		ReceivedAt: time.Now(),
		MessageID:  messageID,
	}
}

// Mock message handler for testing
func createMockHandler(called *bool, shouldError bool) TopicMessageHandler {
	return func(ctx context.Context, message TopicMessage) error {
		*called = true
		if shouldError {
			return errors.New("handler error")
		}
		return nil
	}
}

// Test NewSHIPTopicManager

func TestNewSHIPTopicManager(t *testing.T) {
	mockStorage := new(MockSHIPStorageInterface)

	topicManager := NewSHIPTopicManager(mockStorage, nil)

	assert.NotNil(t, topicManager)
	assert.Equal(t, mockStorage, topicManager.storage)
	assert.Nil(t, topicManager.lookupService)
	assert.NotNil(t, topicManager.subscriptions)
	assert.NotNil(t, topicManager.handlers)
	assert.Equal(t, 0, len(topicManager.subscriptions))
	assert.Equal(t, 0, len(topicManager.handlers))
}

func TestNewSHIPTopicManagerWithLookupService(t *testing.T) {
	topicManager, mockStorage, lookupService := createTestSHIPTopicManagerWithLookupService()

	assert.NotNil(t, topicManager)
	assert.Equal(t, mockStorage, topicManager.storage)
	assert.Equal(t, lookupService, topicManager.lookupService)
	assert.NotNil(t, topicManager.subscriptions)
	assert.NotNil(t, topicManager.handlers)
}

// Test SubscribeToTopic

func TestSubscribeToTopic_Success(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	handlerCalled := false
	handler := createMockHandler(&handlerCalled, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test", handler)

	assert.NoError(t, err)
	assert.True(t, topicManager.IsSubscribedToTopic("tm_test"))
	assert.Equal(t, 1, len(topicManager.subscriptions))
	assert.Equal(t, 1, len(topicManager.handlers))

	// Check subscription details
	subscriptions := topicManager.GetSubscribedTopics()
	assert.Equal(t, 1, len(subscriptions))
	assert.Equal(t, "tm_test", subscriptions[0].Topic)
	assert.True(t, subscriptions[0].IsActive)
	assert.Equal(t, int64(0), subscriptions[0].MessageCount)
}

func TestSubscribeToTopic_EmptyTopic(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	handlerCalled := false
	handler := createMockHandler(&handlerCalled, false)

	err := topicManager.SubscribeToTopic(context.Background(), "", handler)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "topic name cannot be empty")
	assert.False(t, topicManager.IsSubscribedToTopic(""))
}

func TestSubscribeToTopic_NilHandler(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message handler cannot be nil")
	assert.False(t, topicManager.IsSubscribedToTopic("tm_test"))
}

func TestSubscribeToTopic_UpdateExistingSubscription(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Create initial subscription
	handlerCalled1 := false
	handler1 := createMockHandler(&handlerCalled1, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test", handler1)
	assert.NoError(t, err)

	// Unsubscribe
	err = topicManager.UnsubscribeFromTopic(context.Background(), "tm_test")
	assert.NoError(t, err)
	assert.False(t, topicManager.IsSubscribedToTopic("tm_test"))

	// Resubscribe with new handler
	handlerCalled2 := false
	handler2 := createMockHandler(&handlerCalled2, false)

	err = topicManager.SubscribeToTopic(context.Background(), "tm_test", handler2)
	assert.NoError(t, err)
	assert.True(t, topicManager.IsSubscribedToTopic("tm_test"))

	// Should still have only one subscription
	assert.Equal(t, 1, len(topicManager.subscriptions))
	assert.Equal(t, 1, len(topicManager.handlers))
}

// Test UnsubscribeFromTopic

func TestUnsubscribeFromTopic_Success(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// First subscribe
	handlerCalled := false
	handler := createMockHandler(&handlerCalled, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test", handler)
	assert.NoError(t, err)
	assert.True(t, topicManager.IsSubscribedToTopic("tm_test"))

	// Then unsubscribe
	err = topicManager.UnsubscribeFromTopic(context.Background(), "tm_test")
	assert.NoError(t, err)
	assert.False(t, topicManager.IsSubscribedToTopic("tm_test"))

	// Subscription should still exist but be inactive
	assert.Equal(t, 1, len(topicManager.subscriptions))
	assert.Equal(t, 0, len(topicManager.handlers))

	subscriptions := topicManager.GetSubscribedTopics()
	assert.Equal(t, 1, len(subscriptions))
	assert.False(t, subscriptions[0].IsActive)
}

func TestUnsubscribeFromTopic_EmptyTopic(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	err := topicManager.UnsubscribeFromTopic(context.Background(), "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "topic name cannot be empty")
}

func TestUnsubscribeFromTopic_NotSubscribed(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	err := topicManager.UnsubscribeFromTopic(context.Background(), "tm_nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not subscribed to topic: tm_nonexistent")
}

// Test HandleTopicMessage

func TestHandleTopicMessage_Success(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Subscribe to topic
	handlerCalled := false
	handler := createMockHandler(&handlerCalled, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test", handler)
	assert.NoError(t, err)

	// Handle message
	message := createTestTopicMessage("tm_test", "msg-1", "test payload")
	err = topicManager.HandleTopicMessage(context.Background(), message)

	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, int64(1), topicManager.GetTopicMessageCount("tm_test"))
}

func TestHandleTopicMessage_EmptyTopic(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	message := createTestTopicMessage("", "msg-1", "test payload")
	err := topicManager.HandleTopicMessage(context.Background(), message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message topic cannot be empty")
}

func TestHandleTopicMessage_NotSubscribed(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	message := createTestTopicMessage("tm_nonexistent", "msg-1", "test payload")
	err := topicManager.HandleTopicMessage(context.Background(), message)

	// Should silently ignore messages for topics we're not subscribed to
	assert.NoError(t, err)
}

func TestHandleTopicMessage_InactiveSubscription(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Subscribe and then unsubscribe
	handlerCalled := false
	handler := createMockHandler(&handlerCalled, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test", handler)
	assert.NoError(t, err)

	err = topicManager.UnsubscribeFromTopic(context.Background(), "tm_test")
	assert.NoError(t, err)

	// Try to handle message
	message := createTestTopicMessage("tm_test", "msg-1", "test payload")
	err = topicManager.HandleTopicMessage(context.Background(), message)

	// Should silently ignore inactive subscriptions
	assert.NoError(t, err)
	assert.False(t, handlerCalled)
}

func TestHandleTopicMessage_HandlerError(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Subscribe with error handler
	handlerCalled := false
	handler := createMockHandler(&handlerCalled, true)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test", handler)
	assert.NoError(t, err)

	// Handle message
	message := createTestTopicMessage("tm_test", "msg-1", "test payload")
	err = topicManager.HandleTopicMessage(context.Background(), message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to handle message for topic tm_test")
	assert.True(t, handlerCalled)

	// Message count should still be incremented
	assert.Equal(t, int64(1), topicManager.GetTopicMessageCount("tm_test"))
}

// Test CreateTopicSubscription

func TestCreateTopicSubscription_Success(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	subscription, err := topicManager.CreateTopicSubscription(context.Background(), "tm_test")

	assert.NoError(t, err)
	assert.NotNil(t, subscription)
	assert.Equal(t, "tm_test", subscription.Topic)
	assert.False(t, subscription.IsActive) // Should not be active without handler
	assert.Equal(t, int64(0), subscription.MessageCount)
	assert.False(t, topicManager.IsSubscribedToTopic("tm_test"))

	// Should exist in subscriptions
	assert.Equal(t, 1, len(topicManager.subscriptions))
}

func TestCreateTopicSubscription_EmptyTopic(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	subscription, err := topicManager.CreateTopicSubscription(context.Background(), "")

	assert.Error(t, err)
	assert.Nil(t, subscription)
	assert.Contains(t, err.Error(), "topic name cannot be empty")
}

func TestCreateTopicSubscription_ExistingSubscription(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Create first subscription
	subscription1, err := topicManager.CreateTopicSubscription(context.Background(), "tm_test")
	assert.NoError(t, err)

	// Try to create again
	subscription2, err := topicManager.CreateTopicSubscription(context.Background(), "tm_test")
	assert.NoError(t, err)

	// Should return the same subscription
	assert.Equal(t, subscription1, subscription2)
	assert.Equal(t, 1, len(topicManager.subscriptions))
}

// Test GetSubscribedTopics

func TestGetSubscribedTopics_Empty(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	subscriptions := topicManager.GetSubscribedTopics()

	assert.NotNil(t, subscriptions)
	assert.Equal(t, 0, len(subscriptions))
}

func TestGetSubscribedTopics_Multiple(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Create multiple subscriptions
	handlerCalled1 := false
	handler1 := createMockHandler(&handlerCalled1, false)

	handlerCalled2 := false
	handler2 := createMockHandler(&handlerCalled2, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test1", handler1)
	assert.NoError(t, err)

	err = topicManager.SubscribeToTopic(context.Background(), "tm_test2", handler2)
	assert.NoError(t, err)

	// Create an inactive subscription
	_, err = topicManager.CreateTopicSubscription(context.Background(), "tm_test3")
	assert.NoError(t, err)

	subscriptions := topicManager.GetSubscribedTopics()

	assert.Equal(t, 3, len(subscriptions))

	// Count active and inactive
	activeCount := 0
	inactiveCount := 0
	for _, sub := range subscriptions {
		if sub.IsActive {
			activeCount++
		} else {
			inactiveCount++
		}
	}

	assert.Equal(t, 2, activeCount)
	assert.Equal(t, 1, inactiveCount)
}

// Test IsSubscribedToTopic

func TestIsSubscribedToTopic_True(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	handlerCalled := false
	handler := createMockHandler(&handlerCalled, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test", handler)
	assert.NoError(t, err)

	assert.True(t, topicManager.IsSubscribedToTopic("tm_test"))
}

func TestIsSubscribedToTopic_False(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	assert.False(t, topicManager.IsSubscribedToTopic("tm_nonexistent"))
}

func TestIsSubscribedToTopic_Inactive(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Create inactive subscription
	_, err := topicManager.CreateTopicSubscription(context.Background(), "tm_test")
	assert.NoError(t, err)

	assert.False(t, topicManager.IsSubscribedToTopic("tm_test"))
}

// Test GetTopicMessageCount

func TestGetTopicMessageCount_Zero(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	count := topicManager.GetTopicMessageCount("tm_nonexistent")
	assert.Equal(t, int64(0), count)
}

func TestGetTopicMessageCount_WithMessages(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Subscribe and handle messages
	handlerCalled := false
	handler := createMockHandler(&handlerCalled, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test", handler)
	assert.NoError(t, err)

	// Handle multiple messages
	for i := 0; i < 5; i++ {
		message := createTestTopicMessage("tm_test", fmt.Sprintf("msg-%d", i), "test payload")
		err = topicManager.HandleTopicMessage(context.Background(), message)
		assert.NoError(t, err)
	}

	count := topicManager.GetTopicMessageCount("tm_test")
	assert.Equal(t, int64(5), count)
}

// Test Close

func TestClose_Success(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Create multiple subscriptions
	handlerCalled1 := false
	handler1 := createMockHandler(&handlerCalled1, false)

	handlerCalled2 := false
	handler2 := createMockHandler(&handlerCalled2, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test1", handler1)
	assert.NoError(t, err)

	err = topicManager.SubscribeToTopic(context.Background(), "tm_test2", handler2)
	assert.NoError(t, err)

	// Close the topic manager
	err = topicManager.Close(context.Background())
	assert.NoError(t, err)

	// All subscriptions should be inactive
	assert.False(t, topicManager.IsSubscribedToTopic("tm_test1"))
	assert.False(t, topicManager.IsSubscribedToTopic("tm_test2"))

	// All handlers should be cleared
	assert.Equal(t, 0, len(topicManager.handlers))

	// Subscriptions should still exist but be inactive
	subscriptions := topicManager.GetSubscribedTopics()
	assert.Equal(t, 2, len(subscriptions))
	for _, sub := range subscriptions {
		assert.False(t, sub.IsActive)
	}
}

// Test metadata and statistics methods

func TestGetTopicManagerMetaData(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	metadata := topicManager.GetTopicManagerMetaData()

	assert.Equal(t, "SHIP Topic Manager", metadata.Name)
	assert.Equal(t, "Manages overlay network topic subscriptions for SHIP protocol.", metadata.ShortDescription)
}

func TestGetActiveTopicCount(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Initially should be 0
	assert.Equal(t, 0, topicManager.GetActiveTopicCount())

	// Add active subscriptions
	handlerCalled1 := false
	handler1 := createMockHandler(&handlerCalled1, false)

	handlerCalled2 := false
	handler2 := createMockHandler(&handlerCalled2, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test1", handler1)
	assert.NoError(t, err)

	err = topicManager.SubscribeToTopic(context.Background(), "tm_test2", handler2)
	assert.NoError(t, err)

	assert.Equal(t, 2, topicManager.GetActiveTopicCount())

	// Add inactive subscription
	_, err = topicManager.CreateTopicSubscription(context.Background(), "tm_test3")
	assert.NoError(t, err)

	assert.Equal(t, 2, topicManager.GetActiveTopicCount())

	// Unsubscribe from one
	err = topicManager.UnsubscribeFromTopic(context.Background(), "tm_test1")
	assert.NoError(t, err)

	assert.Equal(t, 1, topicManager.GetActiveTopicCount())
}

func TestGetTotalMessageCount(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Initially should be 0
	assert.Equal(t, int64(0), topicManager.GetTotalMessageCount())

	// Subscribe to topics
	handlerCalled1 := false
	handler1 := createMockHandler(&handlerCalled1, false)

	handlerCalled2 := false
	handler2 := createMockHandler(&handlerCalled2, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test1", handler1)
	assert.NoError(t, err)

	err = topicManager.SubscribeToTopic(context.Background(), "tm_test2", handler2)
	assert.NoError(t, err)

	// Handle messages on both topics
	for i := 0; i < 3; i++ {
		message := createTestTopicMessage("tm_test1", fmt.Sprintf("msg-%d", i), "payload")
		err = topicManager.HandleTopicMessage(context.Background(), message)
		assert.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		message := createTestTopicMessage("tm_test2", fmt.Sprintf("msg-%d", i), "payload")
		err = topicManager.HandleTopicMessage(context.Background(), message)
		assert.NoError(t, err)
	}

	assert.Equal(t, int64(5), topicManager.GetTotalMessageCount())
}

// Test concurrent access scenarios

func TestConcurrentSubscription(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Test concurrent subscription to different topics
	done := make(chan bool, 2)

	go func() {
		handlerCalled := false
		handler := createMockHandler(&handlerCalled, false)
		err := topicManager.SubscribeToTopic(context.Background(), "tm_test1", handler)
		assert.NoError(t, err)
		done <- true
	}()

	go func() {
		handlerCalled := false
		handler := createMockHandler(&handlerCalled, false)
		err := topicManager.SubscribeToTopic(context.Background(), "tm_test2", handler)
		assert.NoError(t, err)
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	assert.Equal(t, 2, topicManager.GetActiveTopicCount())
	assert.True(t, topicManager.IsSubscribedToTopic("tm_test1"))
	assert.True(t, topicManager.IsSubscribedToTopic("tm_test2"))
}

func TestConcurrentMessageHandling(t *testing.T) {
	topicManager, _ := createTestSHIPTopicManager()

	// Subscribe to topic
	handlerCalled := false
	handler := createMockHandler(&handlerCalled, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test", handler)
	assert.NoError(t, err)

	// Handle messages concurrently
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(messageID int) {
			message := createTestTopicMessage("tm_test", fmt.Sprintf("msg-%d", messageID), "payload")
			err := topicManager.HandleTopicMessage(context.Background(), message)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	assert.Equal(t, int64(5), topicManager.GetTopicMessageCount("tm_test"))
}

// Integration test with lookup service

func TestIntegrationWithLookupService(t *testing.T) {
	topicManager, mockStorage, lookupService := createTestSHIPTopicManagerWithLookupService()

	// Verify integration
	assert.NotNil(t, topicManager.lookupService)
	assert.Equal(t, lookupService, topicManager.lookupService)

	// Topic manager should still work normally
	handlerCalled := false
	handler := createMockHandler(&handlerCalled, false)

	err := topicManager.SubscribeToTopic(context.Background(), "tm_test", handler)
	assert.NoError(t, err)
	assert.True(t, topicManager.IsSubscribedToTopic("tm_test"))

	// Storage should be the same instance
	assert.Equal(t, mockStorage, topicManager.storage)
}

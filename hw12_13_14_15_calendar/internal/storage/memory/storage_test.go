package memorystorage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/storage"
)

func TestStorage(t *testing.T) {
	s := New()

	t.Run("add", func(t *testing.T) {
		id, err := s.Add(context.TODO(), storage.Event{
			Title:       "Title_1",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(10 * time.Hour),
			Description: "Description_1",
			OwnerID:     1,
			SendBefore:  18000,
		})
		require.NoError(t, err)
		require.Equal(t, uint64(1), id)

		id, err = s.Add(context.TODO(), storage.Event{
			Title:       "Title_2",
			StartAt:     time.Now().Add(1 * time.Hour),
			EndAt:       time.Now().Add(10 * time.Hour),
			Description: "",
			OwnerID:     2,
			SendBefore:  0,
		})
		require.NoError(t, err)
		require.Equal(t, uint64(2), id)
	})

	t.Run("update", func(t *testing.T) {
		err := s.Update(context.TODO(), storage.Event{
			ID:          1,
			Title:       "Title_3",
			StartAt:     time.Now().Add(2 * time.Hour),
			EndAt:       time.Now().Add(10 * time.Hour),
			Description: "Description_3",
			OwnerID:     10,
			SendBefore:  10000,
		})
		require.NoError(t, err)

		err = s.Update(context.TODO(), storage.Event{
			Title:       "Title_2",
			StartAt:     time.Now().Add(1 * time.Hour),
			EndAt:       time.Now().Add(10 * time.Hour),
			Description: "",
			OwnerID:     11,
			SendBefore:  0,
		})
		require.Error(t, err)
	})

	t.Run("list", func(t *testing.T) {
		list, err := s.List(context.TODO(), storage.Params{})
		require.NoError(t, err)
		require.Len(t, list, 2)

		item0 := list[0]
		require.Equal(t, "Title_2", item0.Title)
		require.Equal(t, "", item0.Description)
		require.Equal(t, uint64(2), item0.OwnerID)
		require.Equal(t, int64(0), item0.SendBefore)

		item1 := list[1]
		require.Equal(t, uint64(1), item1.ID)
		require.Equal(t, "Title_3", item1.Title)
		require.Equal(t, "Description_3", item1.Description)
		require.Equal(t, uint64(10), item1.OwnerID)
		require.Equal(t, int64(10000), item1.SendBefore)

		require.True(t, item0.StartAt.Before(item1.StartAt))

		list, err = s.List(context.TODO(), storage.Params{Limit: 1})
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, uint64(2), list[0].ID)

		list, err = s.List(context.TODO(), storage.Params{StartAtGEq: item1.StartAt.Add(-1 * time.Minute)})
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, uint64(1), list[0].ID)
	})

	t.Run("delete", func(t *testing.T) {
		err := s.Delete(context.TODO(), []uint64{1})
		require.NoError(t, err)
		list, err := s.List(context.TODO(), storage.Params{})
		require.NoError(t, err)
		require.Len(t, list, 1)

		err = s.Delete(context.TODO(), []uint64{2})
		require.NoError(t, err)
		list, err = s.List(context.TODO(), storage.Params{})
		require.NoError(t, err)
		require.Len(t, list, 0)

		require.NoError(t, s.Delete(context.TODO(), []uint64{1}))
		require.NoError(t, s.Delete(context.TODO(), []uint64{2}))
	})

	t.Run("ListToSend", func(t *testing.T) {
		id, err := s.Add(context.TODO(), storage.Event{
			StartAt:    time.Now(),
			SendBefore: 60,
		})
		require.NoError(t, err)

		events, err := s.ListToSend(context.TODO())
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Equal(t, id, events[0].ID)
	})

	t.Run("SetSent", func(t *testing.T) {
		err := s.SetSent(context.TODO(), []uint64{100})
		require.NoError(t, err)

		events, err := s.List(context.TODO(), storage.Params{})
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.False(t, events[0].IsSent)

		err = s.SetSent(context.TODO(), []uint64{events[0].ID})
		require.NoError(t, err)

		events, err = s.List(context.TODO(), storage.Params{})
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.True(t, events[0].IsSent)
	})
}

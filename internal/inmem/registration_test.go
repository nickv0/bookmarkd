package inmem_test

import (
	"bookmarkd"
	"errors"
	"testing"

	"bookmarkd/internal/inmem"
	"bookmarkd/utils/require"

	"github.com/google/uuid"
)

var usernames = []string{"user1", "user2", "user3", "user4", "user5"}

func TestRegistrationenticationServiceHandler(t *testing.T) {
	store := inmem.NewRegistrationStore()
	registrationIDs := make([]uuid.UUID, 0)

	t.Run("should start an registration", func(t *testing.T) {

		for _, username := range usernames {
			a, err := store.StartRegistration(username)
			require.Equal(t, err, nil)

			registrationIDs = append(registrationIDs, a.ID)
		}
	})

	t.Run("should return an registration by ID", func(t *testing.T) {

		for i, id := range registrationIDs {
			data, err := store.FindRegistrationByID(id.String())

			require.Equal(t, err, nil)
			require.Equal(t, string(data.Username), string(usernames[i]))
			require.Equal(t, len(data.Seed), 36)
		}
	})

	t.Run("should delete an registration", func(t *testing.T) {

		require.Equal(t, len(registrationIDs), 5)

		err := store.DeleteRegistration(registrationIDs[4].String())
		require.Equal(t, err, nil)

		for i, id := range registrationIDs {
			data, err := store.FindRegistrationByID(id.String())

			if i == 4 {
				t.Log("Looking for 4")
				require.Equal(t, errors.Is(err, bookmarkd.ErrNotFound), true)
			} else {
				require.Equal(t, err, nil)
				require.Equal(t, string(data.Username), string(usernames[i]))
				require.Equal(t, len(data.Seed), 36)
			}
		}

	})

}

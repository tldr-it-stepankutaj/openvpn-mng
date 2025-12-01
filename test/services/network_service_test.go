package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
	"github.com/tldr-it-stepankutaj/openvpn-mng/test/testutil"
)

func TestNetworkService_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewNetworkService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully creates network with CIDR", func(t *testing.T) {
		req := &dto.CreateNetworkRequest{
			Name:        "Server Network",
			CIDR:        "192.168.1.0/24",
			Description: "Internal servers",
		}

		network, err := service.Create(req, admin.ID)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, network.ID)
		assert.Equal(t, "Server Network", network.Name)
		assert.Equal(t, "192.168.1.0/24", network.CIDR)
		assert.Equal(t, "Internal servers", network.Description)
	})

	t.Run("normalizes single IPv4 to CIDR", func(t *testing.T) {
		req := &dto.CreateNetworkRequest{
			Name: "Single IP Network",
			CIDR: "10.0.0.1",
		}

		network, err := service.Create(req, admin.ID)
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.1/32", network.CIDR)
	})

	t.Run("fails with invalid CIDR", func(t *testing.T) {
		req := &dto.CreateNetworkRequest{
			Name: "Invalid Network",
			CIDR: "not-a-valid-cidr",
		}

		_, err := service.Create(req, admin.ID)
		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidCIDR, err)
	})

	t.Run("fails when network name exists", func(t *testing.T) {
		existingNetwork := testutil.CreateTestNetwork(t, admin.ID)

		req := &dto.CreateNetworkRequest{
			Name: existingNetwork.Name,
			CIDR: "10.10.10.0/24",
		}

		_, err := service.Create(req, admin.ID)
		assert.Error(t, err)
		assert.Equal(t, services.ErrNetworkExists, err)
	})

	t.Run("creates network without description", func(t *testing.T) {
		req := &dto.CreateNetworkRequest{
			Name: "Simple Network",
			CIDR: "172.16.0.0/16",
		}

		network, err := service.Create(req, admin.ID)
		require.NoError(t, err)
		assert.Equal(t, "Simple Network", network.Name)
		assert.Empty(t, network.Description)
	})

	t.Run("accepts various CIDR formats", func(t *testing.T) {
		testCases := []struct {
			name         string
			cidr         string
			expectedCIDR string
		}{
			{"Class A", "10.0.0.0/8", "10.0.0.0/8"},
			{"Class B", "172.16.0.0/12", "172.16.0.0/12"},
			{"Class C", "192.168.100.0/24", "192.168.100.0/24"},
			{"Single IP", "8.8.8.8", "8.8.8.8/32"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := &dto.CreateNetworkRequest{
					Name: "Network " + tc.name,
					CIDR: tc.cidr,
				}

				network, err := service.Create(req, admin.ID)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedCIDR, network.CIDR)
			})
		}
	})
}

func TestNetworkService_GetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewNetworkService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully gets network", func(t *testing.T) {
		testNetwork := testutil.CreateTestNetwork(t, admin.ID)

		network, err := service.GetByID(testNetwork.ID)
		require.NoError(t, err)
		assert.Equal(t, testNetwork.ID, network.ID)
		assert.Equal(t, testNetwork.Name, network.Name)
	})

	t.Run("returns error for non-existent network", func(t *testing.T) {
		_, err := service.GetByID(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, services.ErrNetworkNotFound, err)
	})
}

func TestNetworkService_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewNetworkService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully updates network", func(t *testing.T) {
		testNetwork := testutil.CreateTestNetwork(t, admin.ID)

		req := &dto.UpdateNetworkRequest{
			Name:        "Updated Network",
			Description: "Updated description",
		}

		network, err := service.Update(testNetwork.ID, req, admin.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Network", network.Name)
		assert.Equal(t, "Updated description", network.Description)
	})

	t.Run("updates CIDR", func(t *testing.T) {
		testNetwork := testutil.CreateTestNetwork(t, admin.ID)

		req := &dto.UpdateNetworkRequest{
			CIDR: "10.20.30.0/24",
		}

		network, err := service.Update(testNetwork.ID, req, admin.ID)
		require.NoError(t, err)
		assert.Equal(t, "10.20.30.0/24", network.CIDR)
	})

	t.Run("fails with invalid CIDR", func(t *testing.T) {
		testNetwork := testutil.CreateTestNetwork(t, admin.ID)

		req := &dto.UpdateNetworkRequest{
			CIDR: "invalid",
		}

		_, err := service.Update(testNetwork.ID, req, admin.ID)
		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidCIDR, err)
	})

	t.Run("returns error for non-existent network", func(t *testing.T) {
		req := &dto.UpdateNetworkRequest{
			Name: "Test",
		}

		_, err := service.Update(uuid.New(), req, admin.ID)
		assert.Error(t, err)
	})
}

func TestNetworkService_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewNetworkService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully deletes network", func(t *testing.T) {
		testNetwork := testutil.CreateTestNetwork(t, admin.ID)

		err := service.Delete(testNetwork.ID)
		require.NoError(t, err)

		_, err = service.GetByID(testNetwork.ID)
		assert.Error(t, err)
	})

	t.Run("removes group associations on delete", func(t *testing.T) {
		network := testutil.CreateTestNetwork(t, admin.ID)
		group := testutil.CreateTestGroup(t, admin.ID)

		// Add group to network
		service.AddGroupToNetwork(network.ID, group.ID, admin.ID)

		// Delete network
		err := service.Delete(network.ID)
		require.NoError(t, err)

		// Verify network is deleted
		_, err = service.GetByID(network.ID)
		assert.Error(t, err)
	})
}

func TestNetworkService_List(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewNetworkService()
	admin := testutil.CreateTestAdmin(t)

	// Create test networks
	testutil.CreateTestNetwork(t, admin.ID)
	testutil.CreateTestNetwork(t, admin.ID)
	testutil.CreateTestNetwork(t, admin.ID)

	t.Run("lists all networks", func(t *testing.T) {
		networks, total, err := service.List(1, 100)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(networks), 3)
	})

	t.Run("pagination works", func(t *testing.T) {
		networks, _, err := service.List(1, 2)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(networks), 2)
	})
}

func TestNetworkService_GroupManagement(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewNetworkService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("adds group to network", func(t *testing.T) {
		network := testutil.CreateTestNetwork(t, admin.ID)
		group := testutil.CreateTestGroup(t, admin.ID)

		err := service.AddGroupToNetwork(network.ID, group.ID, admin.ID)
		require.NoError(t, err)

		groups, err := service.GetNetworkGroups(network.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(groups))
		assert.Equal(t, group.ID, groups[0].ID)
	})

	t.Run("fails to add duplicate group", func(t *testing.T) {
		network := testutil.CreateTestNetwork(t, admin.ID)
		group := testutil.CreateTestGroup(t, admin.ID)

		err := service.AddGroupToNetwork(network.ID, group.ID, admin.ID)
		require.NoError(t, err)

		err = service.AddGroupToNetwork(network.ID, group.ID, admin.ID)
		assert.Error(t, err)
	})

	t.Run("removes group from network", func(t *testing.T) {
		network := testutil.CreateTestNetwork(t, admin.ID)
		group := testutil.CreateTestGroup(t, admin.ID)

		service.AddGroupToNetwork(network.ID, group.ID, admin.ID)
		err := service.RemoveGroupFromNetwork(network.ID, group.ID)
		require.NoError(t, err)

		groups, err := service.GetNetworkGroups(network.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(groups))
	})

	t.Run("fails for non-existent network", func(t *testing.T) {
		group := testutil.CreateTestGroup(t, admin.ID)

		err := service.AddGroupToNetwork(uuid.New(), group.ID, admin.ID)
		assert.Error(t, err)
	})

	t.Run("fails for non-existent group", func(t *testing.T) {
		network := testutil.CreateTestNetwork(t, admin.ID)

		err := service.AddGroupToNetwork(network.ID, uuid.New(), admin.ID)
		assert.Error(t, err)
	})
}

func TestNetworkService_GetNetworkGroups(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewNetworkService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("returns groups for network", func(t *testing.T) {
		network := testutil.CreateTestNetwork(t, admin.ID)
		group1 := testutil.CreateTestGroup(t, admin.ID)
		group2 := testutil.CreateTestGroup(t, admin.ID)

		service.AddGroupToNetwork(network.ID, group1.ID, admin.ID)
		service.AddGroupToNetwork(network.ID, group2.ID, admin.ID)

		groups, err := service.GetNetworkGroups(network.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(groups))
	})

	t.Run("returns empty for network without groups", func(t *testing.T) {
		network := testutil.CreateTestNetwork(t, admin.ID)

		groups, err := service.GetNetworkGroups(network.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(groups))
	})
}

func TestNetworkService_GetGroupNetworks(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewNetworkService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("returns networks for group", func(t *testing.T) {
		group := testutil.CreateTestGroup(t, admin.ID)
		network1 := testutil.CreateTestNetwork(t, admin.ID)
		network2 := testutil.CreateTestNetwork(t, admin.ID)

		service.AddGroupToNetwork(network1.ID, group.ID, admin.ID)
		service.AddGroupToNetwork(network2.ID, group.ID, admin.ID)

		networks, err := service.GetGroupNetworks(group.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(networks))
	})

	t.Run("returns empty for group without networks", func(t *testing.T) {
		group := testutil.CreateTestGroup(t, admin.ID)

		networks, err := service.GetGroupNetworks(group.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(networks))
	})
}

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

func TestGroupService_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewGroupService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully creates group", func(t *testing.T) {
		req := &dto.CreateGroupRequest{
			Name:        "IT Department",
			Description: "IT team members",
		}

		group, err := service.Create(req, admin.ID)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, group.ID)
		assert.Equal(t, "IT Department", group.Name)
		assert.Equal(t, "IT team members", group.Description)
	})

	t.Run("fails when group name exists", func(t *testing.T) {
		existingGroup := testutil.CreateTestGroup(t, admin.ID)

		req := &dto.CreateGroupRequest{
			Name:        existingGroup.Name,
			Description: "Different description",
		}

		_, err := service.Create(req, admin.ID)
		assert.Error(t, err)
		assert.Equal(t, services.ErrGroupExists, err)
	})

	t.Run("creates group without description", func(t *testing.T) {
		req := &dto.CreateGroupRequest{
			Name: "Simple Group",
		}

		group, err := service.Create(req, admin.ID)
		require.NoError(t, err)
		assert.Equal(t, "Simple Group", group.Name)
		assert.Empty(t, group.Description)
	})
}

func TestGroupService_GetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewGroupService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully gets group", func(t *testing.T) {
		testGroup := testutil.CreateTestGroup(t, admin.ID)

		group, err := service.GetByID(testGroup.ID)
		require.NoError(t, err)
		assert.Equal(t, testGroup.ID, group.ID)
		assert.Equal(t, testGroup.Name, group.Name)
	})

	t.Run("returns error for non-existent group", func(t *testing.T) {
		_, err := service.GetByID(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, services.ErrGroupNotFound, err)
	})
}

func TestGroupService_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewGroupService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully updates group", func(t *testing.T) {
		testGroup := testutil.CreateTestGroup(t, admin.ID)

		req := &dto.UpdateGroupRequest{
			Name:        "Updated Name",
			Description: "Updated description",
		}

		group, err := service.Update(testGroup.ID, req, admin.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", group.Name)
		assert.Equal(t, "Updated description", group.Description)
	})

	t.Run("updates only provided fields", func(t *testing.T) {
		testGroup := testutil.CreateTestGroup(t, admin.ID)
		originalDescription := testGroup.Description

		req := &dto.UpdateGroupRequest{
			Name: "New Name Only",
		}

		group, err := service.Update(testGroup.ID, req, admin.ID)
		require.NoError(t, err)
		assert.Equal(t, "New Name Only", group.Name)
		assert.Equal(t, originalDescription, group.Description)
	})

	t.Run("returns error for non-existent group", func(t *testing.T) {
		req := &dto.UpdateGroupRequest{
			Name: "Test",
		}

		_, err := service.Update(uuid.New(), req, admin.ID)
		assert.Error(t, err)
	})
}

func TestGroupService_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewGroupService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully soft deletes group", func(t *testing.T) {
		testGroup := testutil.CreateTestGroup(t, admin.ID)

		err := service.Delete(testGroup.ID)
		require.NoError(t, err)

		// Group should not be found (soft deleted)
		_, err = service.GetByID(testGroup.ID)
		assert.Error(t, err)
	})
}

func TestGroupService_List(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewGroupService()
	admin := testutil.CreateTestAdmin(t)

	// Create test groups
	testutil.CreateTestGroup(t, admin.ID)
	testutil.CreateTestGroup(t, admin.ID)
	testutil.CreateTestGroup(t, admin.ID)

	t.Run("lists all groups", func(t *testing.T) {
		groups, total, err := service.List(1, 100)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(groups), 3)
	})

	t.Run("pagination works", func(t *testing.T) {
		groups, _, err := service.List(1, 2)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(groups), 2)
	})
}

func TestGroupService_AddUserToGroup(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewGroupService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully adds user to group", func(t *testing.T) {
		group := testutil.CreateTestGroup(t, admin.ID)
		user := testutil.CreateTestRegularUser(t)

		err := service.AddUserToGroup(group.ID, user.ID, admin.ID)
		require.NoError(t, err)

		// Verify user is in group
		users, err := service.GetGroupUsers(group.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(users))
		assert.Equal(t, user.ID, users[0].ID)
	})

	t.Run("fails for non-existent group", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)

		err := service.AddUserToGroup(uuid.New(), user.ID, admin.ID)
		assert.Error(t, err)
	})

	t.Run("fails for non-existent user", func(t *testing.T) {
		group := testutil.CreateTestGroup(t, admin.ID)

		err := service.AddUserToGroup(group.ID, uuid.New(), admin.ID)
		assert.Error(t, err)
	})
}

func TestGroupService_RemoveUserFromGroup(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewGroupService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully removes user from group", func(t *testing.T) {
		group := testutil.CreateTestGroup(t, admin.ID)
		user := testutil.CreateTestRegularUser(t)

		// Add user to group first
		err := service.AddUserToGroup(group.ID, user.ID, admin.ID)
		require.NoError(t, err)

		// Remove user from group
		err = service.RemoveUserFromGroup(group.ID, user.ID)
		require.NoError(t, err)

		// Verify user is no longer in group
		users, err := service.GetGroupUsers(group.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(users))
	})
}

func TestGroupService_GetGroupUsers(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewGroupService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("returns users in group", func(t *testing.T) {
		group := testutil.CreateTestGroup(t, admin.ID)
		user1 := testutil.CreateTestRegularUser(t)
		user2 := testutil.CreateTestRegularUser(t)

		service.AddUserToGroup(group.ID, user1.ID, admin.ID)
		service.AddUserToGroup(group.ID, user2.ID, admin.ID)

		users, err := service.GetGroupUsers(group.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(users))
	})

	t.Run("returns empty for group without users", func(t *testing.T) {
		group := testutil.CreateTestGroup(t, admin.ID)

		users, err := service.GetGroupUsers(group.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(users))
	})
}

func TestGroupService_GetUserGroups(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewGroupService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("returns groups for user", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)
		group1 := testutil.CreateTestGroup(t, admin.ID)
		group2 := testutil.CreateTestGroup(t, admin.ID)

		service.AddUserToGroup(group1.ID, user.ID, admin.ID)
		service.AddUserToGroup(group2.ID, user.ID, admin.ID)

		groups, err := service.GetUserGroups(user.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(groups))
	})

	t.Run("returns empty for user without groups", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)

		groups, err := service.GetUserGroups(user.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(groups))
	})
}

func TestGroupService_NetworkManagement(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewGroupService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("adds network to group", func(t *testing.T) {
		group := testutil.CreateTestGroup(t, admin.ID)
		network := testutil.CreateTestNetwork(t, admin.ID)

		err := service.AddNetworkToGroup(group.ID, network.ID, admin.ID)
		require.NoError(t, err)

		networks, err := service.GetGroupNetworks(group.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(networks))
		assert.Equal(t, network.ID, networks[0].ID)
	})

	t.Run("fails to add duplicate network", func(t *testing.T) {
		group := testutil.CreateTestGroup(t, admin.ID)
		network := testutil.CreateTestNetwork(t, admin.ID)

		err := service.AddNetworkToGroup(group.ID, network.ID, admin.ID)
		require.NoError(t, err)

		err = service.AddNetworkToGroup(group.ID, network.ID, admin.ID)
		assert.Error(t, err)
	})

	t.Run("removes network from group", func(t *testing.T) {
		group := testutil.CreateTestGroup(t, admin.ID)
		network := testutil.CreateTestNetwork(t, admin.ID)

		service.AddNetworkToGroup(group.ID, network.ID, admin.ID)
		err := service.RemoveNetworkFromGroup(group.ID, network.ID)
		require.NoError(t, err)

		networks, err := service.GetGroupNetworks(group.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(networks))
	})

	t.Run("fails for non-existent group", func(t *testing.T) {
		network := testutil.CreateTestNetwork(t, admin.ID)

		err := service.AddNetworkToGroup(uuid.New(), network.ID, admin.ID)
		assert.Error(t, err)
	})
}

func TestGroupService_GetUserGroupsWithNetworks(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewGroupService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("returns groups with networks", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)
		group := testutil.CreateTestGroup(t, admin.ID)
		network := testutil.CreateTestNetwork(t, admin.ID)

		service.AddUserToGroup(group.ID, user.ID, admin.ID)
		service.AddNetworkToGroup(group.ID, network.ID, admin.ID)

		groups, err := service.GetUserGroupsWithNetworks(user.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(groups))
		assert.Equal(t, group.ID, groups[0].ID)
		assert.Equal(t, 1, len(groups[0].Networks))
		assert.Equal(t, network.ID, groups[0].Networks[0].ID)
	})

	t.Run("returns empty for user without groups", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)

		groups, err := service.GetUserGroupsWithNetworks(user.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(groups))
	})
}

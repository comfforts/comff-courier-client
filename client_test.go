package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	comffC "github.com/comfforts/comff-constants"
	api "github.com/comfforts/comff-courier/api/v1"
	"github.com/comfforts/logger"
)

const TEST_DIR = "data"

func TestCourierClient(t *testing.T) {
	logger := logger.NewTestAppLogger(TEST_DIR)

	for scenario, fn := range map[string]func(
		t *testing.T,
		gc Client,
	){
		"courier CRUD, succeeds": testCourierCRUD,
	} {
		t.Run(scenario, func(t *testing.T) {
			gc, teardown := setup(t, logger)
			defer teardown()
			fn(t, gc)
		})
	}

}

func setup(t *testing.T, logger logger.AppLogger) (
	gc Client,
	teardown func(),
) {
	t.Helper()

	gc, err := NewClient(logger, NewDefaultClientOption())
	require.NoError(t, err)

	return gc, func() {
		t.Logf(" TestCourierClient ended, will close courier client")
		err := gc.Close()
		require.NoError(t, err)
	}
}

func testCourierCRUD(t *testing.T, sc Client) {
	t.Helper()

	rqstr, name, org, street, state := "client-courier-crud-test@gmail.com", "Client courier CRUD test", "client-courier-crud-test", "234 Cabrillo Av", "CA"
	acr := api.AddCourierRequest{
		RequestedBy: rqstr,
		Name:        name,
		Org:         org,
		City:        comffC.SANTA_CLARA,
		PostalCode:  comffC.P95051,
		Country:     comffC.US,
		Height:      6,
		Width:       18,
		Depth:       12,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := sc.RegisterCourier(ctx, &acr)
	require.NoError(t, err)
	require.Equal(t, resp.Ok, true)
	require.Equal(t, resp.Courier.Org, acr.Org)
	require.Equal(t, resp.Courier.Name, acr.Name)

	gResp, err := sc.GetCourier(ctx, &api.GetCourierRequest{
		Id: resp.Courier.Id,
	})
	require.NoError(t, err)
	require.Equal(t, gResp.Courier.Id, resp.Courier.Id)

	ucr := api.UpdateCourierRequest{
		Id:          resp.Courier.Id,
		RequestedBy: rqstr,
		Street:      street,
		City:        comffC.SANTA_CLARA,
		State:       state,
		PostalCode:  comffC.P95051,
		Country:     comffC.US,
		Height:      12,
		Width:       24,
		Depth:       12,
	}
	uResp, err := sc.UpdateCourier(ctx, &ucr)
	require.NoError(t, err)
	require.Equal(t, uResp.Courier.Id, resp.Courier.Id)

	dResp, err := sc.DeleteCourier(ctx, &api.DeleteCourierRequest{
		Id: resp.Courier.Id,
	})
	require.NoError(t, err)
	require.Equal(t, true, dResp.Ok)
}

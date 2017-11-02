// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package conn_test

import (
	"context"
	"testing"

	"time"

	"github.com/10gen/mongo-go-driver/yamgo/internal/conntest"
	"github.com/10gen/mongo-go-driver/yamgo/internal/testutil/helpers"
	. "github.com/10gen/mongo-go-driver/yamgo/private/conn"
	"github.com/stretchr/testify/require"
)

func TestCappedProvider_only_allows_max_number_of_connections(t *testing.T) {
	t.Parallel()

	factory := func(_ context.Context) (Connection, error) {
		return &conntest.MockConnection{}, nil
	}

	cappedProvider := CappedProvider(2, factory)

	_, err := cappedProvider(context.Background())
	require.NoError(t, err)

	_, err = cappedProvider(context.Background())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()
	_, err = cappedProvider(ctx)
	require.Error(t, err)
}

func TestCappedProvider_closing_a_connection_releases_a_resource(t *testing.T) {
	t.Parallel()

	factory := func(_ context.Context) (Connection, error) {
		return &conntest.MockConnection{}, nil
	}

	cappedProvider := CappedProvider(2, factory)

	c1, _ := cappedProvider(context.Background())
	_, err := cappedProvider(context.Background())
	require.NoError(t, err)

	go func() {
		time.Sleep(1 * time.Second)
		testhelpers.RequireNoErrorOnClose(t, c1)
	}()
	_, err = cappedProvider(context.Background())
	require.NoError(t, err)
}

package glia_test

// func TestRender(t *testing.T) {
// 	t.Parallel()

// 	m, s := capnp.NewSingleSegmentMessage(nil)
// 	defer m.Release()

// 	res, err := glia.NewRootResult(s)
// 	require.NoError(t, err)

// 	t.Run("expect-fail/status-unset", func(t *testing.T) {
// 		r := func(ctx context.Context, r glia.Result) error {
// 			return nil
// 		}
// 		err = glia.Render(res, req)
// 		require.ErrorIs(t, err, glia.ErrStatusNotSet)
// 	})

// 	t.Run("expect-succeed/status-guestError", func(t *testing.T) {
// 		r := func(ctx context.Context, r glia.Result) error {
// 			r.SetStatus(glia.Result_Status_guestError)
// 			return nil
// 		}
// 		err = glia.Render(res, req)
// 		require.NoError(t, err)
// 	})

// 	t.Run("expect-fail/status-guestError", func(t *testing.T) {
// 		r := func(ctx context.Context, r glia.Result) error {
// 			return errors.New("test")
// 		}
// 		err = glia.Render(res, req)
// 		require.EqualError(t, err, "test")
// 	})
// }

// func TestOk(t *testing.T) {
// 	t.Parallel()

// 	m, s := capnp.NewSingleSegmentMessage(nil)
// 	defer m.Release()

// 	res, err := glia.NewRootResult(s)
// 	require.NoError(t, err)

// 	t.Run("expect-succeed", func(t *testing.T) {
// 		ok := glia.Ok{0x00, 0x01, 0x02, 0x03}
// 		err := glia.Render(context.TODO(), res, ok)
// 		require.NoError(t, err)

// 		// check status
// 		assert.Equal(t, glia.Result_Status_ok, res.Status())

// 		t.Run("stack", func(t *testing.T) {
// 			stack, err := res.Stack()
// 			require.NoError(t, err)

// 			for i := 0; i < stack.Len(); i++ {
// 				assert.Equal(t, ok[i], stack.At(i))
// 			}
// 		})
// 	})
// }

package workspace

//func TestFindWorkspaces(t *testing.T) {
//	got, err := findWorkspaces([]Module{
//		{"testdata/configs/envs/dev", true},
//		{"testdata/configs/envs/prod", true},
//		{"testdata/configs/envs/staging", true},
//		{"testdata/configs/uninitialized", false},
//	})
//	require.NoError(t, err)
//	assert.Equal(t, 5, len(got))
//	assert.Contains(t, got, workspace{"default", Module{"testdata/configs/envs/dev", true}})
//	assert.Contains(t, got, workspace{"non-default-1", Module{"testdata/configs/envs/dev", true}})
//	assert.Contains(t, got, workspace{"non-default-2", Module{"testdata/configs/envs/dev", true}})
//	assert.Contains(t, got, workspace{"default", Module{"testdata/configs/envs/prod", true}})
//	assert.Contains(t, got, workspace{"default", Module{"testdata/configs/envs/staging", true}})
//}

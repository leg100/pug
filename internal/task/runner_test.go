package task

//func TestRunner_queue(t *testing.T) {
//	// create runner with max number of tasks set to 3, and fire off 3 tasks
//	runner := NewRunner(3, "../testdata/killme")
//	batch1 := make([]*Task, 3)
//	for i := 0; i < 3; i++ {
//		task, err := runner.Run(Spec{})
//		require.NoError(t, err)
//		// wait for it to enter running state
//		assert.Equal(t, Running, <-task.Events, task.Err)
//		batch1[i] = task
//	}
//	// start further batch of 3 tasks, which will be queued
//	batch2 := make([]*Task, 3)
//	for i := 0; i < 3; i++ {
//		task, err := runner.Run(Spec{})
//		require.NoError(t, err)
//		batch2[i] = task
//	}
//	// kill first batch
//	for i := 0; i < 3; i++ {
//		batch1[i].cancel()
//		assert.Equal(t, Errored, <-batch1[i].Events)
//	}
//	// second batch should now run
//	for i := 0; i < 3; i++ {
//		// wait for it to enter running state
//		assert.Equal(t, Running, <-batch2[i].Events)
//	}
//}
//
//func TestRunner_exclusive(t *testing.T) {
//	runner := NewRunner(0, "../testdata/killme")
//	require.Equal(t, 0, len(runner.exclusive))
//
//	// run an exclusive task
//	task, err := runner.Run(Spec{Exclusive: true})
//	require.NoError(t, err)
//	// wait for it to enter running state
//	assert.Equal(t, Running, <-task.Events, task.Err)
//
//	// should have taken exclusive slot
//	require.Equal(t, 1, len(runner.exclusive))
//
//	// start another exclusive task; should be queued
//	another, err := runner.Run(Spec{Exclusive: true})
//	require.NoError(t, err)
//
//	// kill running task
//	task.cancel()
//	assert.Equal(t, Errored, <-task.Events)
//
//	// other task should now run
//	assert.Equal(t, Running, <-another.Events)
//}

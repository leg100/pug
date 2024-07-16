dependency "test_module_a" {
  config_path = "../test_module_a"
}

inputs = {
  time_to_wait = dependency.test_module_a.outputs.time_waited
}

package server

func setupProvider(setup *SetupRequest) error {
	if setup.Provider == "aws" {
		return setupAws(setup)
	}

	return nil
}
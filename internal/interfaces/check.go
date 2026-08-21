package interfaces

import (
	"maksec/internal/adapter/httpserver"
	"maksec/internal/deploy"
	"maksec/internal/repo"
	"maksec/internal/usecase"
)

var (
	_ usecase.IRepoEvents  = (*repo.Events)(nil)
	_ usecase.IRepoScripts = (*repo.Scripts)(nil)

	_ usecase.IDepoyer = (*deploy.Deployer)(nil)

	_ httpserver.IScriptsUseCase  = (*usecase.Scripts)(nil)
	_ httpserver.ICallbackUseCase = (*usecase.Callback)(nil)
)

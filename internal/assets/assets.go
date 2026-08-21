package assets

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed templates/*.sh
var templatesFS embed.FS

//go:embed bin/agent
var agentFS embed.FS

// Templates возвращает файловую систему шаблонов без префикса каталога:
// имена файлов совпадают с именами шаблонов из запроса (template1, template2).
func Templates() fs.FS {
	sub, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		// Недостижимо: каталог зашит в embed на этапе компиляции.
		panic(fmt.Sprintf("assets: templates subdir: %v", err))
	}
	return sub
}

// AgentBinary возвращает содержимое бинарника агента (linux/amd64),
// который deployer кладёт на целевой хост.
func AgentBinary() ([]byte, error) {
	return agentFS.ReadFile("bin/agent")
}

package scripthelpers

func GetHelperMap() map[string]interface{} {
	return map[string]interface{}{
		"request": httpRequest,
	}
}

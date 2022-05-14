package queue2

import metadatabox "github.com/sinmetalcraft/gcpbox/metadata"

func ProjectID() string {
	pID, err := metadatabox.ProjectID()
	if err != nil {
		panic(err)
	}
	return pID
}

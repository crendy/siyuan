package av

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/88250/gulu"
	"github.com/siyuan-note/filelock"
	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/siyuan/kernel/util"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	attributeViewRelationsLock = sync.Mutex{}
)

func GetSrcAvIDs(destAvID string) []string {
	attributeViewRelationsLock.Lock()
	defer attributeViewRelationsLock.Unlock()

	relations := filepath.Join(util.DataDir, "storage", "av", "relations.msgpack")
	if !filelock.IsExist(relations) {
		return nil
	}

	data, err := filelock.ReadFile(relations)
	if nil != err {
		logging.LogErrorf("read attribute view relations failed: %s", err)
		return nil
	}

	avRels := map[string][]string{}
	if err = msgpack.Unmarshal(data, &avRels); nil != err {
		logging.LogErrorf("unmarshal attribute view relations failed: %s", err)
		return nil
	}

	srcAvIDs := avRels[destAvID]
	if nil == srcAvIDs {
		return nil
	}
	return srcAvIDs
}

func RemoveAvRel(srcAvID, destAvID string) {
	attributeViewRelationsLock.Lock()
	defer attributeViewRelationsLock.Unlock()

	relations := filepath.Join(util.DataDir, "storage", "av", "relations.msgpack")
	if !filelock.IsExist(relations) {
		return
	}

	data, err := filelock.ReadFile(relations)
	if nil != err {
		logging.LogErrorf("read attribute view relations failed: %s", err)
		return
	}

	avRels := map[string][]string{}
	if err = msgpack.Unmarshal(data, &avRels); nil != err {
		logging.LogErrorf("unmarshal attribute view relations failed: %s", err)
		return
	}

	srcAvIDs := avRels[destAvID]
	if nil == srcAvIDs {
		return
	}

	var newAvIDs []string
	for _, v := range srcAvIDs {
		if v != srcAvID {
			newAvIDs = append(newAvIDs, v)
		}
	}
	avRels[destAvID] = newAvIDs

	data, err = msgpack.Marshal(avRels)
	if nil != err {
		logging.LogErrorf("marshal attribute view relations failed: %s", err)
		return
	}
	if err = filelock.WriteFile(relations, data); nil != err {
		logging.LogErrorf("write attribute view relations failed: %s", err)
		return
	}
}

func UpsertAvBackRel(srcAvID, destAvID string) {
	attributeViewRelationsLock.Lock()
	defer attributeViewRelationsLock.Unlock()

	avRelations := map[string][]string{}
	relations := filepath.Join(util.DataDir, "storage", "av", "relations.msgpack")
	if !filelock.IsExist(relations) {
		if err := os.MkdirAll(filepath.Dir(relations), 0755); nil != err {
			logging.LogErrorf("create attribute view dir failed: %s", err)
			return
		}
	} else {
		data, err := filelock.ReadFile(relations)
		if nil != err {
			logging.LogErrorf("read attribute view relations failed: %s", err)
			return
		}

		if err = msgpack.Unmarshal(data, &avRelations); nil != err {
			logging.LogErrorf("unmarshal attribute view relations failed: %s", err)
			return
		}
	}

	srcAvIDs := avRelations[destAvID]
	srcAvIDs = append(srcAvIDs, srcAvID)
	srcAvIDs = gulu.Str.RemoveDuplicatedElem(srcAvIDs)
	avRelations[destAvID] = srcAvIDs

	data, err := msgpack.Marshal(avRelations)
	if nil != err {
		logging.LogErrorf("marshal attribute view relations failed: %s", err)
		return
	}
	if err = filelock.WriteFile(relations, data); nil != err {
		logging.LogErrorf("write attribute view relations failed: %s", err)
		return
	}
}

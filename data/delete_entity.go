package data

//// 删除无用的实体
//// 只删除已知无用实体
//func deleteEntity(entities []*Entity) {
//	d := GetDevices()
//	if len(d) > 0 {
//		if len(d) < len(entities)*5 {
//			return
//		}
//	}
//
//	for _, e := range entities {
//		var isDelete bool
//		if e.Category == CategoryLight || e.Category == CategoryLightGroup {
//			if strings.Contains(e.OriginalName, "开关状态切换") {
//				isDelete = true
//			}
//			if strings.Contains()
//		}
//	}
//
//}
//
//func deleteEntityFunc(entityId string) {
//	x.Del(ava.Background(), GetHassUrl()+"/api/states/"+entityId, GetToken(), nil)
//}

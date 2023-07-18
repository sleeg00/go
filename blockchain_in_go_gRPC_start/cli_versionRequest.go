package main

import "log"

func versionRequest(nodeID string) {

	var im int32

	bc, err := GetBlockchain(nodeID) //내 블록체인 가져옴
	if err != nil {
		im = -1
	} else {
		im = int32(bc.GetBestHeight())
	}
	bc.db.Close()
	var copyNode string
	var check bool
	check = false
	log.Println("여기까지 도착")
	for _, node := range knownNodes {
		log.Println(node+"", nodeID)
		if node != nodeID {
			req := sendVersion(node, int64(im))

			if req.GetHeight() > im {
				im = req.GetHeight() //서버연결하고 자신이 아닌 노드에게 길이 요청

				copyNode = node
				//가장 높은 길이를 가진 노드의 번호를 가져옴
				check = true
			}

		}
	}
	if check {
		log.Println("바뀌었음")
		ChangeBlockchain(copyNode, nodeID) //가장 긴 블럭을 자신의 노드로 만듬
		//만약 첫 서버 시작으로 파일이 없다면 생성하고 복사함
	}
}

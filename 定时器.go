
// 防踩坑知识点之 time.After
// https://www.jianshu.com/p/179f405ad2b1

func testTimer() {

	var timer *time.Timer = nil
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
  
  
 	if timer == nil {
		timer = time.NewTimer(timeDuration)
	} else {
		timer.Reset(timeDuration)
	}
  
  select {
		case <- timer.C:
				// 超时了

    case :			// 从管道中读取到了客户端的数据
		
  }

}





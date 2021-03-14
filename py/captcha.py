import cv2
import numpy as np
import base64
from fastapi import FastAPI
import uvicorn as uvicorn
from pydantic import BaseModel

app = FastAPI()
class Captcha(BaseModel):
    dragon: str
    tiger: str

@app.post('/captcha')
def captcha(c: Captcha):
    tl = captcha_fuck(c.dragon,c.tiger)
    return {'success': True, 'x': tl[0]}

def captcha_fuck(bg,tp):
    # bg_img = cv2.imread(bg) # 背景图片
    # tp_img = cv2.imread(tp) # 缺口图片

    bg_img0 = np.fromstring(base64.b64decode(bg), dtype=np.uint8)
    tp_img0 = np.fromstring(base64.b64decode(tp), dtype=np.uint8)
    bg_img = cv2.imdecode(bg_img0, flags=cv2.IMREAD_COLOR)
    tp_img = cv2.imdecode(tp_img0, flags=cv2.IMREAD_COLOR)

    tp_img = tp_img[30:77,0:47]

    bg_edge = cv2.Canny(bg_img, 50, 150)
    tp_edge = cv2.Canny(tp_img, 50, 150)

    bg_pic = cv2.cvtColor(bg_edge, cv2.COLOR_GRAY2RGB)
    tp_pic = cv2.cvtColor(tp_edge, cv2.COLOR_GRAY2RGB)

    res = cv2.matchTemplate(bg_pic, tp_pic, cv2.TM_CCORR_NORMED)
    min_val, max_val, min_loc, max_loc = cv2.minMaxLoc(res)
    # print(min_val, max_val, min_loc, max_loc)

    # th, tw = tp_pic.shape[:2]
    # tl = max_loc #
    # br = (tl[0]+tw,tl[1]+th)
    # cv2.rectangle(bg_img, tl, br, (0, 0, 255), 2)
    # cv2.imwrite("./1.png", bg_edge)
    # cv2.imwrite("./2.png", tp_edge)
    return max_loc


if __name__ == '__main__':
    uvicorn.run(app=app, host="127.0.0.1", port=8080)




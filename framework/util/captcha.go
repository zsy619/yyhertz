package util

import (
	"image/color"
	"sync"

	"github.com/mojocn/base64Captcha"
)

var (
	// 验证码字符集合
	codeSet = "2345678abcdefhijkmnpqrstuvwxyzABCDEFGHJKLMNPQRTUVWXY"

	// 中文验证码字符集
	zhSet = "们以我到他会作时要动国产的一是工就年阶义发成部民可出能方进在了不和有大这主中人上为来分生对于学下级地个用同行面说种过命度革而多子后自社加小机也经力线本电高量长党得实家定深法表着水理化争现所二起政三好十战无农使性前等反体合斗路图把结第里正新开论之物从当两些还天资事队批点育重其思与间内去因件日利相由压员气业代全组数果期导平各基或月毛然如应形想制心样干都向变关问比展那它最及外没看治提五解系林者米群头意只明四道马认次文通但条较克又公孔领军流入接席位情运器并飞原油放立题质指建区验活众很教决特此常石强极土少已根共直团统式转别造切九你取西持总料连任志观调七么山程百报更见必真保热委手改管处己将修支识病象几先老光专什六型具示复安带每东增则完风回南广劳轮科北打积车计给节做务被整联步类集号列温装即毫知轴研单色坚据速防史拉世设达尔场织历花受求传口断况采精金界品判参层止边清至万确究书术状厂须离再目海交权且儿青才证低越际八试规斯近注办布门铁需走议县兵固除般引齿千胜细影济白格效置推空配刀叶率述今选养德话查差半敌始片施响收华觉备名红续均药标记难存测士身紧液派准斤角降维板许破述技消底床田势端感往神便贺村构照容非搞亚磨族火段算适讲按值美态黄易彪服早班麦削信排台声该击素张密害侯草何树肥继右属市严径螺检左页抗苏显苦英快称坏移约巴材省黑武培著河帝仅针怎植京助升王眼她抓含苗副杂普谈围食射源例致酸旧却充足短划剂宣环落首尺波承粉践府鱼随考刻靠够满夫失包住促枝局菌杆周护岩师举曲春元超负砂封换太模贫减阳扬江析亩木言球朝医校古呢稻宋听唯输滑站另卫字鼓刚写刘微略范供阿块某功套友限项余倒卷创律雨让骨远帮初皮播优占死毒圈伟季训控激找叫云互跟裂粮粒母练塞钢顶策双留误础吸阻故寸盾晚丝女散焊功株亲院冷彻弹错散商视艺灭版烈零室轻血倍缺厘泵察绝富城冲喷壤简否柱李望盘磁雄似困巩益洲脱投送奴侧润盖挥距触星松送获兴独官混纪依未突架宽冬章湿偏纹吃执阀矿寨责熟稳夺硬价努翻奇甲预职评读背协损棉侵灰虽矛厚罗泥辟告卵箱掌氧恩爱停曾溶营终纲孟钱待尽俄缩沙退陈讨奋械载胞幼哪剥迫旋征槽倒握担仍呀鲜吧卡粗介钻逐弱脚怕盐末阴丰雾冠丙街莱贝辐肠付吉渗瑞惊顿挤秒悬姆烂森糖圣凹陶词迟蚕亿矩康遵牧遭幅园腔订香肉弟屋敏恢忘编印蜂急拿扩伤飞露核缘游振操央伍域甚迅辉异序免纸夜乡久隶缸夹念兰映沟乙吗儒杀汽磷艰晶插埃燃欢铁补咱芽永瓦倾阵碳演威附牙芽永瓦斜灌欧献顺猪洋腐请透司危括脉宜笑若尾束壮暴企菜穗楚汉愈绿拖牛份染既秋遍锻玉夏疗尖殖井费州访吹荣铜沿替滚客召旱悟刺脑措贯藏敢令隙炉壳硫煤迎铸粘探临薄旬善福纵择礼愿伏残雷延烟句纯渐耕跑泽慢栽鲁赤繁境潮横掉锥希池败船假亮谓托伙哲怀割摆贡呈劲财仪沉炼麻罪祖息车穿货销齐鼠抽画饲龙库守筑房歌寒喜哥洗蚀废纳腹乎录镜妇恶脂庄擦险赞钟摇典柄辩竹谷卖乱虚桥奥伯赶垂途额壁网截野遗静谋弄挂课镇妄盛耐援扎虑键归符庆聚绕摩忙舞遇索顾胶羊湖钉仁音迹碎伸灯避泛亡答勇频皇柳哈揭甘诺概宪浓岛袭谁洪谢炮浇斑讯懂灵蛋闭孩释乳巨徒私银伊景坦累匀霉杜乐勒隔弯绩招绍胡呼痛峰零柴簧午跳居尚丁秦稍追梁折耗碱殊岗挖氏刃剧堆赫荷胸衡勤膜篇登驻案刊秧缓凸役剪川雪链渔啦脸户洛孢勃盟买杨宗焦赛旗滤硅炭股坐蒸凝竟陷枪黎救冒暗洞犯筒您宋弧爆谬涂味津臂障褐陆啊健尊豆拔莫抵桑坡缝警挑污冰柬嘴啥饭塑寄赵喊垫丹渡耳刨虎笔稀昆浪萨茶滴浅拥穴覆伦娘吨浸袖珠雌妈紫戏塔锤震岁貌洁剖牢锋疑霸闪埔猛诉刷狠忽灾闹乔唐漏闻沈熔氯荒茎男凡抢像浆旁玻亦忠唱蒙予纷捕锁尤乘乌智淡允叛畜俘摸锈扫毕璃宝芯爷鉴秘净蒋钙肩腾枯抛轨堂拌爸循诱祝励肯酒绳穷塘燥泡袋朗喂铝软渠颗惯贸粪综墙趋彼届墨碍启逆卸航衣孙龄岭骗休借"

	store = base64Captcha.DefaultMemStore
)

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	Height     int    // 验证码图片高度
	Width      int    // 验证码图片宽度
	Length     int    // 验证码位数
	Source     string // 字符源
	NoiseCount int    // 干扰点数量
	FontSize   int    // 字体大小
}

// DefaultConfig 默认配置
var DefaultConfig = &CaptchaConfig{
	Height:     40,
	Width:      120,
	Length:     5,
	Source:     codeSet,
	NoiseCount: 20,
	FontSize:   25,
}

// CaptchaManager 验证码管理器
type CaptchaManager struct {
	store base64Captcha.Store
	mutex sync.RWMutex
}

// NewCaptchaManager 创建验证码管理器
func NewCaptchaManager() *CaptchaManager {
	return &CaptchaManager{
		store: base64Captcha.DefaultMemStore,
	}
}

// GenerateString 生成字符串验证码
func (c *CaptchaManager) GenerateString(config ...*CaptchaConfig) (id, b64s, answer string, err error) {
	cfg := DefaultConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}

	driver := base64Captcha.DriverString{
		Height:          cfg.Height,
		Width:           cfg.Width,
		NoiseCount:      cfg.NoiseCount,
		ShowLineOptions: base64Captcha.OptionShowHollowLine,
		Length:          cfg.Length,
		Source:          cfg.Source,
		BgColor: &color.RGBA{
			R: 240,
			G: 240,
			B: 246,
			A: 246,
		},
		Fonts: []string{"wqy-microhei.ttc"},
	}

	captcha := base64Captcha.NewCaptcha(driver.ConvertFonts(), c.store)
	return captcha.Generate()
}

// GenerateDigit 生成数字验证码
func (c *CaptchaManager) GenerateDigit(config ...*CaptchaConfig) (id, b64s, answer string, err error) {
	cfg := DefaultConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}

	driver := base64Captcha.DriverDigit{
		Height:   cfg.Height,
		Width:    cfg.Width,
		Length:   cfg.Length,
		MaxSkew:  0.7,
		DotCount: cfg.NoiseCount,
	}

	captcha := base64Captcha.NewCaptcha(&driver, c.store)
	return captcha.Generate()
}

// GenerateMath 生成数学运算验证码
func (c *CaptchaManager) GenerateMath(config ...*CaptchaConfig) (id, b64s, answer string, err error) {
	cfg := DefaultConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}

	driver := base64Captcha.DriverMath{
		Height:          cfg.Height,
		Width:           cfg.Width,
		NoiseCount:      cfg.NoiseCount,
		ShowLineOptions: base64Captcha.OptionShowHollowLine,
		BgColor: &color.RGBA{
			R: 240,
			G: 240,
			B: 246,
			A: 246,
		},
		Fonts: []string{"wqy-microhei.ttc"},
	}

	captcha := base64Captcha.NewCaptcha(driver.ConvertFonts(), c.store)
	return captcha.Generate()
}

// GenerateChinese 生成中文验证码
func (c *CaptchaManager) GenerateChinese(config ...*CaptchaConfig) (id, b64s, answer string, err error) {
	cfg := DefaultConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}

	driver := base64Captcha.DriverChinese{
		Height:          cfg.Height,
		Width:           cfg.Width,
		NoiseCount:      cfg.NoiseCount,
		ShowLineOptions: base64Captcha.OptionShowHollowLine,
		Length:          cfg.Length,
		Source:          zhSet,
		BgColor: &color.RGBA{
			R: 240,
			G: 240,
			B: 246,
			A: 246,
		},
		Fonts: []string{"wqy-microhei.ttc"},
	}

	captcha := base64Captcha.NewCaptcha(driver.ConvertFonts(), c.store)
	return captcha.Generate()
}

// GenerateAudio 生成音频验证码
func (c *CaptchaManager) GenerateAudio(language string) (id, b64s, answer string, err error) {
	if language == "" {
		language = "zh"
	}

	driver := base64Captcha.DriverAudio{
		Length:   6,
		Language: language,
	}

	captcha := base64Captcha.NewCaptcha(&driver, c.store)
	return captcha.Generate()
}

// Verify 验证验证码
func (c *CaptchaManager) Verify(id, answer string, clear bool) bool {
	return c.store.Verify(id, answer, clear)
}

// SetCustomStore 设置自定义存储
func (c *CaptchaManager) SetCustomStore(customStore base64Captcha.Store) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.store = customStore
}

// GetStore 获取存储实例
func (c *CaptchaManager) GetStore() base64Captcha.Store {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.store
}

// 全局验证码管理器实例
var DefaultCaptcha = NewCaptchaManager()

// 便捷函数
func GenerateString(config ...*CaptchaConfig) (id, b64s, answer string, err error) {
	return DefaultCaptcha.GenerateString(config...)
}

func GenerateDigit(config ...*CaptchaConfig) (id, b64s, answer string, err error) {
	return DefaultCaptcha.GenerateDigit(config...)
}

func GenerateMath(config ...*CaptchaConfig) (id, b64s, answer string, err error) {
	return DefaultCaptcha.GenerateMath(config...)
}

func GenerateChinese(config ...*CaptchaConfig) (id, b64s, answer string, err error) {
	return DefaultCaptcha.GenerateChinese(config...)
}

func GenerateAudio(language string) (id, b64s, answer string, err error) {
	return DefaultCaptcha.GenerateAudio(language)
}

func Verify(id, answer string, clear bool) bool {
	return DefaultCaptcha.Verify(id, answer, clear)
}

# Ban-PCDN_Anti-P2P
光复网路计画 (Cn-PCDN-Rules) 防止流氓软件偷走你的宽带上行！本项目致力于搜集国内各大互联网厂商（如爱优腾、B站、抖音快手百度等）用于 P/M/HCDN  偷跑上行的域名、IP 和特征词。通过在网关、DNS 或防火墙层面进行拦截，彻底阻断恶意 P2P 流量，保障家庭核心网络的稳定，避免因异常上行被运营商限速或封号。方便宽带网络爱好者使用。

“Network Liberation Plan" is a specialized collection of rules designed to stop mainstream apps from "stealing" your home's upload bandwidth.

<div style="border-top: 1px solid #d0d7de; margin: 24px 0;"></div>

<div align="center">
  <h1><span style="color:#C0392B; font-weight:bold;">🌐 光复网络计划 (Network Restoration Project)</span></h1>
  <p>防止流氓软件偷走你的宽带上行！拒绝被白嫖，捍卫物理 BRAS 与公网 IP 的主权。</p>
  <p>
    作者：<b>哥哥科技</b> 
    (<a href="https://space.bilibili.com/501430041/" style="color:#FB7299; text-decoration:none;">B站</a> · 
    <a href="https://www.zhihu.com/people/guka0930/posts" style="color:#0084FF; text-decoration:none;">知乎</a>) 
    &nbsp;|&nbsp; 
    QQ 交流群：<a href="https://qun.qq.com/universal-share/share?ac=1&authKey=%2FWJ%2F4NS6lNp1cyU2a1kUo%2F%2FEfq8vD6q7SXUux2UosjZ%2BialqZVNe%2Bk3w1YoOUR3b&busi_data=eyJncm91cENvZGUiOiI2ODA0NjQzNjUiLCJ0b2tlbiI6IkI5ZHVYaHJnSGlZZEpzSkRhamdOZFZaNy9aZFJ1L1JwNGpZYWNJWTJJRThGYjBoVjN6RUxhUkcvV0h4eHBwV2siLCJ1aW4iOiIxNTIwMDg1MjQ5In0%3D&data=iHLL1DGOGIhddoSRjGmI9D2bMbNLa4vwx5YfZQAzvPM0sAuYo67cPN4265_UAPSpbgqWLVl8vmWBZVREf_Qg-g&svctype=4&tempid=h5_group_info" style="color:#1296db; text-decoration:none;">680464365</a> 
    &nbsp;|&nbsp; 
    最新版本：v4.2.1 (2026.04.13)
  </p>
</div>

---

### <span style="color:#D35400;">📖 核心理念：大道至简，力大砖飞</span>

在这个万物皆可 PCDN（端到端内容分发网络）、连你的智能电视和后台驻留软件都在偷偷上传数据吸血的时代，我们必须夺回属于自己的网络带宽。本项目的核心目的，就是通过<span style="color:#27AE60; font-weight:bold;">最底层、最纯粹的网络层与 DNS 层拦截</span>，彻底斩断那些潜伏在暗处的恶意 P2P/PCDN 偷跑行为。

作为有线网络原教旨主义倡导者，我们坚决抵制为了屏蔽几个域名，部署基于 Arm 架构的“软路由”或繁冗的应用层插件、小主机或各类花里胡哨的插件。**真正的极客，追求的是硬路由的 NPU 线速转发！** 用最暴力的防火墙策略，守护最纯净的 Open Internet。不搞流控，不搞 QoS，把所有的算力留给真正的业务，这才是美式的“量大管饱”。

---

### <span style="color:#8E44AD;">🚀 项目组成与拦截策略</span>

可以根据自己家里的路由器性能（建议使用支持高级 URL/IP 过滤的硬路由）和实际业务需求进行组合部署：

#### 1. 核心关键词与域名拦截 (`PCDN域名.csv`)
基于 URL/关键字的黑名单阻断方案，适用于支持三层关键字过滤的硬路由。规则集按拦截置信度进行了分级：
* **<span style="color:#2C3E50;">黑色加粗（高置信度）：</span>** 纯坏、已明确确认为恶意 PCDN 的流氓域名（如 `bsccdn`, `yfp2p`, `ppio`, `黑心云` 等）。如果你追求稳定，只加这些即可！**建议默认全量部署**。
* **<span style="color:#2980B9;">蓝色标注（极客注意）：</span>** 可能会误杀部分极客业务（如部分特定 P2P 隧道），按需屏蔽。可能干预部分私有或极客，需网络管理员根据本地业务按需启用。
* **<span style="color:#C0392B;">红色标注（高危测试）：</span>** 处于推测阶段的特征值，极易引发大面积业务异常，非特定调试环境请勿部署。
* **<span style="color:#F39C12;">黄色底色（低置信度）：</span>** 激进策略，极大概率误伤普通应用的正常加载。

> <span style="color:#7F8C8D;">*小贴士：能在应用层解决的问题（比如用“李跳跳”去广告），就不要在三层网络层折腾。手机端看 B站建议直接上第三方客户端（如 PiliPlus），电脑端永远只用 Web 网页版并配合篡改猴插件！*</span>

**推荐插件**：WebRTC Control，Global Speed，B站ACG助手，pakku弹幕过滤

Make bilibili Great Again，BiliBili 高级倍速功能，【哔哩哔哩】屏蔽视频PCDN地址，Bilibili Evolved (Preview)，更好的哔哩哔哩

bilibili播放视频倍速可随意拖动 1x-4x 手动输入/滑块控制/微调按钮，哔哩哔哩网页版显示 IP 属地 B站 Bilibili IP 属地显示

#### 2. Hosts 沉洞与精准 IP 路由 (`Hosts.csv`)
将已知的流氓调度节点直接指向回环地址或无效地址：
* 拦截 `ppio.cloud`, `yfcdn.net`, `onethingpcs.com` 等重灾区，强制解析至 `127.114.9.80` 或 `0.0.0.0`。
* 屏蔽特定厂商的静默上传。

#### 3. 广告与流媒体 PCDN 净化 (`ADGHome.csv`)
为 AdGuard Home 维护者整合的合并过滤规则集：不仅抑制 PCDN，还能兼顾去广告，同时设置了严格的<span style="color:#27AE60;">基础服务白名单</span>，防止长辈在家使用时图片加载失败。
涵盖的流媒体重灾区：腾讯视频 (`hcdn.video.qq.com`)、爱奇艺、优酷土豆、抖音/西瓜/字节系 (`flive.douyincdn.com`)。

#### 4. 海外纯净防漏策略 (`海外.csv`)
专为防止“忘开代理”导致的 DNS 泄露与流量特征暴露。在 UDP 53 层面直接拦截特定海外服务商（Google, ChatGPT, 广告联盟等）的直连解析，避免你的宝贵全球唯一单播地址被重点关注。

---

### <span style="color:#16A085;">🛠️ 实施示例 (以中兴ZTE路由器为例)</span>

正如 `光复网路计划初版.pdf` 中展示的那样，硬路由的配置才是男人的浪漫。
1. 登录你的中兴路由器后台（或其他企业级/运营商级设备）。
2. 导航至 **安全 -> 防火墙 -> URL 过滤**。
3. 将模式设置为 **“黑名单”**。
4. 依次将 `PCDN域名.csv` 中**黑色加粗**的关键词（如 `pcdn`, `mcdn`, `hcdn`, `p2p`）添加到 URL 过滤规则中。
5. 提交并重启相关服务，享受不被白嫖的满血上行宽带！

*(可选) 进阶网维可参考 `对称型防火墙（进阶）.csv`，研究无状态防火墙下的回程放行逻辑与 TCP 8082 端口的拦截策略。*

---

### <span style="color:#2C3E50;">🔄 更新日志 (v4.2.1 - 2026.04.13 21H UTC+8)</span>

* **第 16 次修正**：根据群友与长辈反馈，基于多端设备测试反馈，回调了上个版本中过于激进的自用拦截策略。
* 通过 PCAPdroid 对移动端应用进行了重新嗅探抓包，大幅降低了常规软件的网络卡顿率。
* 优化了分类色彩逻辑：目前所有激进部分均已用黄色标出。
* 新增关于 Bilibili MCDN（TCP 8082）的拦截评估。建议在客户端层面规避（如 PC 端使用 WebRTC Control 扩展，移动端使用 PiliPlus 第三方客户端），而非单纯依赖网关拦截。
* **核心准则更新**：如果遇到网络问题排查，优先删除自己添加的其他规则，**黑色加粗**的安全底线无需删除。如果你的路由器比较高级，可以查看具体的连接拦截日志，欢迎进群联合排查。

---

<div align="center">
  <p><span style="color:#E74C3C; font-size:1.1em; font-weight:bold;">“在一个文明社会，干净的、不被监视与吸血的网络，是我们每个人的基本权利。”</span></p>
  <p>如果你也是有线原教旨主义者，如果你也热爱干净的路由表与专线级体验，欢迎加入我们的阵营。
    <br>点击 Star，加入光复网路的行列吧！</p>
</div>

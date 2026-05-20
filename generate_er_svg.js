const fs = require('fs');
const path = require('path');

// 系统ER图的SVG (Chen记法)
const systemERSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1200 950" width="1200" height="950">
  <defs>
    <style>
      .entity { fill: #ffffff; stroke: #000000; stroke-width: 2.5px; }
      .entity-text { font-family: 'SimSun', serif; font-size: 16px; font-weight: bold; fill: #000000; }
      .relation { fill: #ffffff; stroke: #000000; stroke-width: 1.5px; }
      .relation-text { font-family: 'SimSun', serif; font-size: 13px; fill: #000000; }
      .line { stroke: #000000; stroke-width: 1.5px; fill: none; }
      .cardinality { font-family: 'SimSun', serif; font-size: 15px; font-weight: bold; fill: #000000; }
    </style>
  </defs>

  <!-- 标题 -->
  <text x="600" y="30" text-anchor="middle" style="font-family: SimSun, serif; font-size: 22px; font-weight: bold;">系统E-R图</text>

  <!-- 用户实体 (矩形) -->
  <rect x="480" y="420" width="100" height="50" class="entity"/>
  <text x="530" y="452" text-anchor="middle" class="entity-text">用户</text>

  <!-- 项目实体 (矩形) -->
  <rect x="200" y="150" width="100" height="50" class="entity"/>
  <text x="250" y="182" text-anchor="middle" class="entity-text">项目</text>

  <!-- 集群实体 (矩形) -->
  <rect x="850" y="150" width="100" height="50" class="entity"/>
  <text x="900" y="182" text-anchor="middle" class="entity-text">集群</text>

  <!-- 操作日志实体 (矩形) -->
  <rect x="520" y="750" width="100" height="50" class="entity"/>
  <text x="570" y="782" text-anchor="middle" class="entity-text">操作日志</text>

  <!-- 关系：创建 (菱形) -->
  <polygon points="365,310 445,310 405,340 365,310" class="relation" transform="translate(0, -10)"/>
  <text x="405" y="335" text-anchor="middle" class="relation-text">创建</text>

  <!-- 关系：配置 (菱形) -->
  <polygon points="860,310 940,310 900,340 860,310" class="relation" transform="translate(0, -10)"/>
  <text x="900" y="335" text-anchor="middle" class="relation-text">配置</text>

  <!-- 关系：产生 (菱形) -->
  <polygon points="550,630 630,630 590,660 550,630" class="relation" transform="translate(0, -10)"/>
  <text x="590" y="655" text-anchor="middle" class="relation-text">产生</text>

  <!-- 连接线：用户-创建 -->
  <line x1="505" y1="420" x2="390" y2="330" class="line"/>
  <text x="485" y="395" class="cardinality">1</text>

  <!-- 连接线：创建-项目 -->
  <line x1="405" y1="300" x2="270" y2="200" class="line"/>
  <text x="255" y="245" class="cardinality">N</text>

  <!-- 连接线：用户-配置 -->
  <line x1="555" y1="420" x2="880" y2="330" class="line"/>
  <text x="585" y="395" class="cardinality">1</text>

  <!-- 连接线：配置-集群 -->
  <line x1="900" y1="300" x2="900" y2="200" class="line"/>
  <text x="905" y="245" class="cardinality">N</text>

  <!-- 连接线：用户-产生 -->
  <line x1="530" y1="470" x2="570" y2="620" class="line"/>
  <text x="538" y="520" class="cardinality">1</text>

  <!-- 连接线：产生-操作日志 -->
  <line x1="590" y1="650" x2="570" y2="750" class="line"/>
  <text x="578" y="720" class="cardinality">N</text>

</svg>`;

// 用户实体属性图的SVG (辐射式)
const userEntitySVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 900 700" width="900" height="700">
  <defs>
    <style>
      .entity { fill: #ffffff; stroke: #000000; stroke-width: 2.5px; }
      .entity-text { font-family: 'SimSun', serif; font-size: 16px; font-weight: bold; fill: #000000; }
      .attribute { fill: #ffffff; stroke: #000000; stroke-width: 1.5px; }
      .attr-text { font-family: 'SimSun', serif; font-size: 12px; fill: #000000; }
      .attr-pk { font-weight: bold; text-decoration: underline; }
      .line { stroke: #000000; stroke-width: 1.2px; fill: none; }
    </style>
  </defs>

  <!-- 标题 -->
  <text x="450" y="30" text-anchor="middle" style="font-family: SimSun, serif; font-size: 22px; font-weight: bold;">用户实体属性图</text>

  <!-- 用户表 (矩形 - 中心) -->
  <rect x="380" y="320" width="100" height="50" class="entity"/>
  <text x="430" y="352" text-anchor="middle" class="entity-text">用户表</text>

  <!-- 属性：id (椭圆) - 主键 -->
  <ellipse cx="115" cy="440" rx="45" ry="22" class="attribute"/>
  <text x="115" y="445" text-anchor="middle" class="attr-text attr-pk">id</text>
  <line x1="160" y1="440" x2="380" y2="345" class="line"/>

  <!-- 属性：用户名 (椭圆) -->
  <ellipse cx="120" cy="320" rx="50" ry="22" class="attribute"/>
  <text x="120" y="325" text-anchor="middle" class="attr-text">用户名</text>
  <line x1="170" y1="320" x2="380" y2="338" class="line"/>

  <!-- 属性：邮箱 (椭圆) -->
  <ellipse cx="95" cy="200" rx="45" ry="22" class="attribute"/>
  <text x="95" y="205" text-anchor="middle" class="attr-text">邮箱</text>
  <line x1="130" y1="218" x2="400" y2="320" class="line"/>

  <!-- 属性：角色类型 (椭圆) -->
  <ellipse cx="240" cy="110" rx="55" ry="22" class="attribute"/>
  <text x="240" y="115" text-anchor="middle" class="attr-text">角色类型</text>
  <line x1="265" y1="132" x2="415" y2="320" class="line"/>

  <!-- 属性：密码哈希 (椭圆) -->
  <ellipse cx="430" cy="70" rx="55" ry="22" class="attribute"/>
  <text x="430" y="75" text-anchor="middle" class="attr-text">密码哈希</text>
  <line x1="430" y1="92" x2="430" y2="320" class="line"/>

  <!-- 属性：创建时间 (椭圆) -->
  <ellipse cx="620" cy="110" rx="55" ry="22" class="attribute"/>
  <text x="620" y="115" text-anchor="middle" class="attr-text">创建时间</text>
  <line x1="595" y1="132" x2="445" y2="320" class="line"/>

  <!-- 属性：更新时间 (椭圆) -->
  <ellipse cx="760" cy="200" rx="55" ry="22" class="attribute"/>
  <text x="760" y="205" text-anchor="middle" class="attr-text">更新时间</text>
  <line x1="725" y1="218" x2="460" y2="320" class="line"/>

</svg>`;

// 项目实体属性图的SVG (辐射式)
const projectEntitySVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1000 800" width="1000" height="800">
  <defs>
    <style>
      .entity { fill: #ffffff; stroke: #000000; stroke-width: 2.5px; }
      .entity-text { font-family: 'SimSun', serif; font-size: 16px; font-weight: bold; fill: #000000; }
      .attribute { fill: #ffffff; stroke: #000000; stroke-width: 1.5px; }
      .attr-text { font-family: 'SimSun', serif; font-size: 12px; fill: #000000; }
      .attr-pk { font-weight: bold; text-decoration: underline; }
      .line { stroke: #000000; stroke-width: 1.2px; fill: none; }
    </style>
  </defs>

  <!-- 标题 -->
  <text x="500" y="30" text-anchor="middle" style="font-family: SimSun, serif; font-size: 22px; font-weight: bold;">项目实体属性图</text>

  <!-- 项目表 (矩形 - 中心) -->
  <rect x="420" y="360" width="100" height="50" class="entity"/>
  <text x="470" y="392" text-anchor="middle" class="entity-text">项目表</text>

  <!-- 属性：id (椭圆) - 主键 -->
  <ellipse cx="115" cy="480" rx="42" ry="22" class="attribute"/>
  <text x="115" y="485" text-anchor="middle" class="attr-text attr-pk">id</text>
  <line x1="157" y1="472" x2="420" y2="390" class="line"/>

  <!-- 属性：项目名称 (椭圆) -->
  <ellipse cx="105" cy="350" rx="52" ry="22" class="attribute"/>
  <text x="105" y="355" text-anchor="middle" class="attr-text">项目名称</text>
  <line x1="157" y1="350" x2="420" y2="372" class="line"/>

  <!-- 属性：项目描述 (椭圆) -->
  <ellipse cx="75" cy="220" rx="52" ry="22" class="attribute"/>
  <text x="75" y="225" text-anchor="middle" class="attr-text">项目描述</text>
  <line x1="110" y1="242" x2="430" y2="360" class="line"/>

  <!-- 属性：所属用户ID (椭圆) - 外键 -->
  <ellipse cx="230" cy="122" rx="58" ry="25" class="attribute"/>
  <text x="230" y="127" text-anchor="middle" class="attr-text" style="font-size:11px;">所属用户ID</text>
  <line x1="260" y1="147" x2="440" y2="360" class="line"/>

  <!-- 属性：数据库配置 (椭圆) -->
  <ellipse cx="430" cy="62" rx="58" ry="25" class="attribute"/>
  <text x="430" y="67" text-anchor="middle" class="attr-text" style="font-size:11px;">数据库配置</text>
  <line x1="430" y1="87" x2="460" y2="360" class="line"/>

  <!-- 属性：表结构配置 (椭圆) -->
  <ellipse cx="620" cy="82" rx="58" ry="25" class="attribute"/>
  <text x="620" y="87" text-anchor="middle" class="attr-text" style="font-size:11px;">表结构配置</text>
  <line x1="590" y1="107" x2="480" y2="360" class="line"/>

  <!-- 属性：生成代码 (椭圆) -->
  <ellipse cx="805" cy="170" rx="52" ry="22" class="attribute"/>
  <text x="805" y="175" text-anchor="middle" class="attr-text">生成代码</text>
  <line x1="758" y1="182" x2="520" y2="368" class="line"/>

  <!-- 属性：创建时间 (椭圆) -->
  <ellipse cx="865" cy="300" rx="52" ry="22" class="attribute"/>
  <text x="865" y="305" text-anchor="middle" class="attr-text">创建时间</text>
  <line x1="820" y1="315" x2="520" y2="378" class="line"/>

  <!-- 属性：更新时间 (椭圆) -->
  <ellipse cx="845" cy="420" rx="52" ry="22" class="attribute"/>
  <text x="845" y="425" text-anchor="middle" class="attr-text">更新时间</text>
  <line x1="793" y1="420" x2="520" y2="395" class="line"/>

</svg>`;

// 集群实体属性图的SVG (辐射式)
const clusterEntitySVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1100 900" width="1100" height="900">
  <defs>
    <style>
      .entity { fill: #ffffff; stroke: #000000; stroke-width: 2.5px; }
      .entity-text { font-family: 'SimSun', serif; font-size: 16px; font-weight: bold; fill: #000000; }
      .attribute { fill: #ffffff; stroke: #000000; stroke-width: 1.5px; }
      .attr-text { font-family: 'SimSun', serif; font-size: 12px; fill: #000000; }
      .attr-pk { font-weight: bold; text-decoration: underline; }
      .line { stroke: #000000; stroke-width: 1.2px; fill: none; }
    </style>
  </defs>

  <!-- 标题 -->
  <text x="550" y="30" text-anchor="middle" style="font-family: SimSun, serif; font-size: 22px; font-weight: bold;">集群实体属性图</text>

  <!-- 集群表 (矩形 - 中心) -->
  <rect x="480" y="420" width="100" height="50" class="entity"/>
  <text x="530" y="452" text-anchor="middle" class="entity-text">集群表</text>

  <!-- 属性：id (椭圆) - 主键 -->
  <ellipse cx="135" cy="540" rx="42" ry="22" class="attribute"/>
  <text x="135" y="545" text-anchor="middle" class="attr-text attr-pk">id</text>
  <line x1="177" y1="532" x2="480" y2="450" class="line"/>

  <!-- 属性：集群名称 (椭圆) -->
  <ellipse cx="105" cy="400" rx="52" ry="22" class="attribute"/>
  <text x="105" y="405" text-anchor="middle" class="attr-text">集群名称</text>
  <line x1="157" y1="400" x2="480" y2="432" class="line"/>

  <!-- 属性：集群描述 (椭圆) -->
  <ellipse cx="75" cy="260" rx="52" ry="22" class="attribute"/>
  <text x="75" y="265" text-anchor="middle" class="attr-text">集群描述</text>
  <line x1="110" y1="282" x2="495" y2="420" class="line"/>

  <!-- 属性：所属用户ID (椭圆) - 外键 -->
  <ellipse cx="230" cy="152" rx="58" ry="25" class="attribute"/>
  <text x="230" y="157" text-anchor="middle" class="attr-text" style="font-size:11px;">所属用户ID</text>
  <line x1="260" y1="177" x2="505" y2="420" class="line"/>

  <!-- 属性：Docker地址 (椭圆) -->
  <ellipse cx="407" cy="72" rx="55" ry="25" class="attribute"/>
  <text x="407" y="77" text-anchor="middle" class="attr-text" style="font-size:11px;">Docker地址</text>
  <line x1="430" y1="97" x2="520" y2="420" class="line"/>

  <!-- 属性：K8s API地址 (椭圆) -->
  <ellipse cx="600" cy="52" rx="58" ry="25" class="attribute"/>
  <text x="600" y="57" text-anchor="middle" class="attr-text" style="font-size:11px;">K8s API地址</text>
  <line x1="575" y1="77" x2="540" y2="420" class="line"/>

  <!-- 属性：集群类型 (椭圆) -->
  <ellipse cx="785" cy="100" rx="52" ry="22" class="attribute"/>
  <text x="785" y="105" text-anchor="middle" class="attr-text">集群类型</text>
  <line x1="748" y1="118" x2="565" y2="422" class="line"/>

  <!-- 属性：状态 (椭圆) -->
  <ellipse cx="905" cy="200" rx="42" ry="22" class="attribute"/>
  <text x="905" y="205" text-anchor="middle" class="attr-text">状态</text>
  <line x1="863" y1="210" x2="580" y2="432" class="line"/>

  <!-- 属性：版本号 (椭圆) -->
  <ellipse cx="940" cy="320" rx="47" ry="22" class="attribute"/>
  <text x="940" y="325" text-anchor="middle" class="attr-text">版本号</text>
  <line x1="893" y1="325" x2="580" y2="445" class="line"/>

  <!-- 属性：节点数量 (椭圆) -->
  <ellipse cx="925" cy="440" rx="52" ry="22" class="attribute"/>
  <text x="925" y="445" text-anchor="middle" class="attr-text">节点数量</text>
  <line x1="873" y1="442" x2="580" y2="450" class="line"/>

  <!-- 属性：KubeConfig (椭圆) -->
  <ellipse cx="900" cy="552" rx="58" ry="25" class="attribute"/>
  <text x="900" y="557" text-anchor="middle" class="attr-text" style="font-size:11px;">KubeConfig</text>
  <line x1="842" y1="548" x2="580" y2="462" class="line"/>

</svg>`;

// 操作日志实体属性图的SVG (辐射式)
const logEntitySVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1200 1000" width="1200" height="1000">
  <defs>
    <style>
      .entity { fill: #ffffff; stroke: #000000; stroke-width: 2.5px; }
      .entity-text { font-family: 'SimSun', serif; font-size: 16px; font-weight: bold; fill: #000000; }
      .attribute { fill: #ffffff; stroke: #000000; stroke-width: 1.5px; }
      .attr-text { font-family: 'SimSun', serif; font-size: 12px; fill: #000000; }
      .attr-pk { font-weight: bold; text-decoration: underline; }
      .line { stroke: #000000; stroke-width: 1.2px; fill: none; }
    </style>
  </defs>

  <!-- 标题 -->
  <text x="600" y="30" text-anchor="middle" style="font-family: SimSun, serif; font-size: 22px; font-weight: bold;">操作日志实体属性图</text>

  <!-- 操作日志表 (矩形 - 中心) -->
  <rect x="530" y="480" width="120" height="50" class="entity"/>
  <text x="590" y="512" text-anchor="middle" class="entity-text">操作日志表</text>

  <!-- 属性：id (椭圆) - 主键 -->
  <ellipse cx="135" cy="620" rx="42" ry="22" class="attribute"/>
  <text x="135" y="625" text-anchor="middle" class="attr-text attr-pk">id</text>
  <line x1="177" y1="612" x2="530" y2="510" class="line"/>

  <!-- 属性：操作用户ID (椭圆) - 外键 -->
  <ellipse cx="92" cy="482" rx="58" ry="25" class="attribute"/>
  <text x="92" y="487" text-anchor="middle" class="attr-text" style="font-size:11px;">操作用户ID</text>
  <line x1="150" y1="485" x2="530" y2="500" class="line"/>

  <!-- 属性：操作用户名 (椭圆) -->
  <ellipse cx="72" cy="342" rx="58" ry="25" class="attribute"/>
  <text x="72" y="347" text-anchor="middle" class="attr-text" style="font-size:11px;">操作用户名</text>
  <line x1="118" y1="367" x2="540" y2="480" class="line"/>

  <!-- 属性：操作类型 (椭圆) -->
  <ellipse cx="125" cy="200" rx="52" ry="22" class="attribute"/>
  <text x="125" y="205" text-anchor="middle" class="attr-text">操作类型</text>
  <line x1="158" y1="222" x2="555" y2="480" class="line"/>

  <!-- 属性：资源类型 (椭圆) -->
  <ellipse cx="295" cy="90" rx="52" ry="22" class="attribute"/>
  <text x="295" y="95" text-anchor="middle" class="attr-text">资源类型</text>
  <line x1="320" y1="112" x2="568" y2="480" class="line"/>

  <!-- 属性：资源ID (椭圆) -->
  <ellipse cx="492" cy="50" rx="47" ry="22" class="attribute"/>
  <text x="492" y="55" text-anchor="middle" class="attr-text">资源ID</text>
  <line x1="508" y1="72" x2="578" y2="480" class="line"/>

  <!-- 属性：操作详情 (椭圆) -->
  <ellipse cx="695" cy="45" rx="52" ry="22" class="attribute"/>
  <text x="695" y="50" text-anchor="middle" class="attr-text">操作详情</text>
  <line x1="670" y1="67" x2="600" y2="480" class="line"/>

  <!-- 属性：操作状态 (椭圆) -->
  <ellipse cx="895" cy="80" rx="52" ry="22" class="attribute"/>
  <text x="895" y="85" text-anchor="middle" class="attr-text">操作状态</text>
  <line x1="860" y1="102" x2="625" y2="482" class="line"/>

  <!-- 属性：客户端IP (椭圆) -->
  <ellipse cx="1055" cy="170" rx="52" ry="22" class="attribute"/>
  <text x="1055" y="175" text-anchor="middle" class="attr-text">客户端IP</text>
  <line x1="1008" y1="182" x2="650" y2="492" class="line"/>

  <!-- 属性：操作耗时 (椭圆) -->
  <ellipse cx="1095" cy="302" rx="62" ry="25" class="attribute"/>
  <text x="1095" y="307" text-anchor="middle" class="attr-text" style="font-size:10px;">操作耗时(ms)</text>
  <line x1="1033" y1="305" x2="650" y2="502" class="line"/>

  <!-- 属性：错误信息 (椭圆) -->
  <ellipse cx="1065" cy="420" rx="52" ry="22" class="attribute"/>
  <text x="1065" y="425" text-anchor="middle" class="attr-text">错误信息</text>
  <line x1="1013" y1="422" x2="650" y2="505" class="line"/>

  <!-- 属性：操作时间 (椭圆) -->
  <ellipse cx="1005" cy="540" rx="52" ry="22" class="attribute"/>
  <text x="1005" y="545" text-anchor="middle" class="attr-text">操作时间</text>
  <line x1="953" y1="538" x2="650" y2="518" class="line"/>

</svg>`;

// 保存SVG文件
const diagramsDir = './diagrams';

fs.writeFileSync(path.join(diagramsDir, '4-3-1-系统ER图.svg'), systemERSVG);
console.log('✓ 已更新: 4-3-1-系统ER图.svg');

fs.writeFileSync(path.join(diagramsDir, '4-3-2-用户实体属性图.svg'), userEntitySVG);
console.log('✓ 已更新: 4-3-2-用户实体属性图.svg');

fs.writeFileSync(path.join(diagramsDir, '4-3-3-项目实体属性图.svg'), projectEntitySVG);
console.log('✓ 已更新: 4-3-3-项目实体属性图.svg');

fs.writeFileSync(path.join(diagramsDir, '4-3-4-集群实体属性图.svg'), clusterEntitySVG);
console.log('✓ 已更新: 4-3-4-集群实体属性图.svg');

fs.writeFileSync(path.join(diagramsDir, '4-3-5-操作日志实体属性图.svg'), logEntitySVG);
console.log('✓ 已更新: 4-3-5-操作日志实体属性图.svg');

console.log('\n所有ER图已按照新样式重新生成！');

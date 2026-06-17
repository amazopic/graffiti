const D=JSON.parse(document.getElementById('graffiti-data').textContent);
const N=D.label.length,E=D.edges,cat=D.cat,file=D.file;
const cv=document.getElementById('c'),ctx=cv.getContext('2d');
let W,H,DPR=Math.min(2,devicePixelRatio||1);
function resize(){W=innerWidth;H=innerHeight;cv.width=W*DPR;cv.height=H*DPR;cv.style.width=W+'px';cv.style.height=H+'px';ctx.setTransform(DPR,0,0,DPR,0,0);}
addEventListener('resize',resize);resize();
const px=new Float64Array(N),py=new Float64Array(N);
const GA=Math.PI*(3-Math.sqrt(5));for(let i=0;i<N;i++){const r=18*Math.sqrt(i+1),a=i*GA;px[i]=Math.cos(a)*r;py[i]=Math.sin(a)*r;}
function dirOf(f){const i=f.lastIndexOf('/');return i<0?'·':f.slice(0,i);}
function hue(s){let h=0;for(let i=0;i<s.length;i++)h=(h*31+s.charCodeAt(i))>>>0;return h%360;}
function colorFor(dir){return `hsl(${hue(dir)} 66% 62%)`;}
const sectorOf=file.map(dirOf),EXT='#8b949e';
const elev=new Float32Array(N);
const sphereCache={};
function shade(c){return c.replace('66% 62%','58% 30%');}
function sphere(c){if(sphereCache[c])return sphereCache[c];const s=64,o=document.createElement('canvas');o.width=o.height=s;const g=o.getContext('2d');const grd=g.createRadialGradient(s*0.37,s*0.33,s*0.04,s*0.5,s*0.5,s*0.5);grd.addColorStop(0,'#ffffff');grd.addColorStop(0.16,c);grd.addColorStop(1,shade(c));g.fillStyle=grd;g.beginPath();g.arc(s/2,s/2,s/2-1,0,7);g.fill();return sphereCache[c]=o;}
const adj=Array.from({length:N},()=>[]);for(const[a,b]of E){adj[a].push(b);adj[b].push(a);}
const catOn=[true,true,true];const hidden=new Set();let showZones=true,threeD=true;
function isHidden(f){let path='';for(const p of f.split('/')){path=path?path+'/'+p:p;if(hidden.has(path))return true;}return false;}
function visible(i){return catOn[cat[i]]&&!isHidden(file[i]);}
const K=Math.max(30,420/Math.sqrt(Math.max(1,N)));let temp=K*5,alive=true,iter=0;const MAXITER=400;let dragNode=-1;
function step(){const fx=new Float64Array(N),fy=new Float64Array(N);
  for(let i=0;i<N;i++)for(let j=i+1;j<N;j++){let dx=px[i]-px[j],dy=py[i]-py[j];let d2=dx*dx+dy*dy;if(d2<0.01){d2=0.01;dx=(i-j)*0.1+0.1;}const f=K*K/d2,d=Math.sqrt(d2),ux=dx/d,uy=dy/d;fx[i]+=ux*f;fy[i]+=uy*f;fx[j]-=ux*f;fy[j]-=uy*f;}
  for(const[a,b]of E){let dx=px[a]-px[b],dy=py[a]-py[b];const d=Math.hypot(dx,dy)||0.01,f=d*d/K,ux=dx/d,uy=dy/d;fx[a]-=ux*f;fy[a]-=uy*f;fx[b]+=ux*f;fy[b]+=uy*f;}
  for(let i=0;i<N;i++){fx[i]-=px[i]*0.02;fy[i]-=py[i]*0.02;}
  for(let i=0;i<N;i++){if(i===dragNode)continue;const d=Math.hypot(fx[i],fy[i])||0.01,m=Math.min(d,temp);px[i]+=fx[i]/d*m;py[i]+=fy[i]/d*m;}
  temp*=0.985;iter++;if(iter>=MAXITER||temp<0.6)alive=false;}
let scale=1,ox=0,oy=0,fitted=false;
function fitArr(ax,ay){let a=1e9,b=1e9,c=-1e9,e=-1e9;for(let i=0;i<N;i++){a=Math.min(a,ax[i]);b=Math.min(b,ay[i]);c=Math.max(c,ax[i]);e=Math.max(e,ay[i]);}const gw=c-a||1,gh=e-b||1;scale=0.88*Math.min(W/gw,H/gh);ox=W/2-((a+c)/2)*scale;oy=H/2-((b+e)/2)*scale;}
function fitVisible(){let a=1e9,b=1e9,c=-1e9,e=-1e9,any=false;for(let i=0;i<N;i++){if(!visible(i))continue;any=true;a=Math.min(a,px[i]);b=Math.min(b,py[i]);c=Math.max(c,px[i]);e=Math.max(e,py[i]);}if(!any)return;const gw=c-a||1,gh=e-b||1;scale=Math.min(40,0.86*Math.min(W/gw,H/gh));ox=W/2-((a+c)/2)*scale;oy=H/2-((b+e)/2)*scale;}
const SX=x=>x*scale+ox,SY=y=>y*scale+oy,WX=s=>(s-ox)/scale,WY=s=>(s-oy)/scale;
function radius(i){return 2.2+Math.sqrt(D.deg[i])*1.4;}
function hull(P){if(P.length<3)return P;P=P.slice().sort((u,v)=>u[0]-v[0]||u[1]-v[1]);const cr=(o,a,b)=>(a[0]-o[0])*(b[1]-o[1])-(a[1]-o[1])*(b[0]-o[0]);const lo=[],up=[];for(const p of P){while(lo.length>1&&cr(lo[lo.length-2],lo[lo.length-1],p)<=0)lo.pop();lo.push(p);}for(let i=P.length-1;i>=0;i--){const p=P[i];while(up.length>1&&cr(up[up.length-2],up[up.length-1],p)<=0)up.pop();up.push(p);}lo.pop();up.pop();return lo.concat(up);}
const bySector={};for(let i=0;i<N;i++){if(cat[i]===2)continue;(bySector[sectorOf[i]]||(bySector[sectorOf[i]]=[])).push(i);}
let hover=-1;const near=new Set();
function draw(){ctx.clearRect(0,0,W,H);
  near.clear();if(hover>=0){near.add(hover);for(const n of adj[hover])near.add(n);}
  for(let i=0;i<N;i++){const tgt=(!threeD||hover<0)?0:(i===hover?28:(near.has(i)?14:0));elev[i]+=(tgt-elev[i])*0.22;}
  if(showZones)for(const s in bySector){const mem=bySector[s].filter(visible);if(mem.length<3)continue;let cx=0,cy=0;for(const i of mem){cx+=px[i];cy+=py[i];}cx/=mem.length;cy/=mem.length;const hp=hull(mem.map(i=>[px[i],py[i]])).map(([x,y])=>{const dx=x-cx,dy=y-cy,L=Math.hypot(dx,dy)||1,o=16/scale;return[SX(x+dx/L*o),SY(y+dy/L*o)];});if(hp.length<3)continue;ctx.beginPath();ctx.moveTo(hp[0][0],hp[0][1]);for(let k=1;k<hp.length;k++)ctx.lineTo(hp[k][0],hp[k][1]);ctx.closePath();ctx.fillStyle=colorFor(s).replace('hsl','hsla').replace(')',' / .08)');ctx.fill();ctx.strokeStyle=colorFor(s).replace('hsl','hsla').replace(')',' / .18)');ctx.lineWidth=1;ctx.stroke();}
  for(const[a,b]of E){if(!visible(a)||!visible(b))continue;const la=elev[a],lb=elev[b],hot=hover>=0&&(a===hover||b===hover);const x1=SX(px[a]),y1=SY(py[a])-la,x2=SX(px[b]),y2=SY(py[b])-lb;
    if(la>0.4||lb>0.4){const mx=(x1+x2)/2,my=(y1+y2)/2-Math.max(la,lb)*0.7;ctx.strokeStyle=hot?'rgba(140,190,255,.6)':'rgba(150,160,190,.22)';ctx.lineWidth=hot?1.1:0.5;ctx.beginPath();ctx.moveTo(x1,y1);ctx.quadraticCurveTo(mx,my,x2,y2);ctx.stroke();}
    else{ctx.strokeStyle='rgba(120,135,160,.07)';ctx.lineWidth=Math.max(0.12,scale*0.16);ctx.beginPath();ctx.moveTo(x1,y1);ctx.lineTo(x2,y2);ctx.stroke();}}
  for(let i=0;i<N;i++)if(!(hover>=0&&near.has(i)))drawNode(i);
  if(hover>=0)for(const i of near)drawNode(i);
  ctx.globalAlpha=1;}
function drawNode(i){if(!visible(i))return;const r=radius(i)*Math.max(.6,Math.min(1.8,scale)),el=elev[i],gx=SX(px[i]),gy=SY(py[i]),y=gy-el,dim=hover>=0&&!near.has(i);ctx.globalAlpha=dim?.16:1;
  if(el>0.4){ctx.globalAlpha=dim?.04:.26;ctx.fillStyle='#000';ctx.beginPath();ctx.ellipse(gx,gy+r*0.35,r*1.05,r*0.45,0,0,7);ctx.fill();ctx.globalAlpha=dim?.16:1;}
  if(cat[i]===2){ctx.strokeStyle=EXT;ctx.lineWidth=1.4;ctx.beginPath();ctx.arc(gx,y,r,0,7);ctx.stroke();}
  else{if(threeD){if(i===hover){ctx.shadowColor='rgba(120,170,255,.9)';ctx.shadowBlur=16;}ctx.drawImage(sphere(colorFor(sectorOf[i])),gx-r,y-r,2*r,2*r);ctx.shadowBlur=0;}else{ctx.fillStyle=colorFor(sectorOf[i]);ctx.beginPath();ctx.arc(gx,y,r,0,7);ctx.fill();}if(cat[i]===1){ctx.globalAlpha=dim?.16:.92;ctx.fillStyle='#0b0f17';ctx.beginPath();ctx.arc(gx,y,r*0.42,0,7);ctx.fill();ctx.globalAlpha=dim?.16:1;}}
  if(i===hover){ctx.strokeStyle='#fff';ctx.lineWidth=1.5;ctx.beginPath();ctx.arc(gx,y,r,0,7);ctx.stroke();}
  if((D.deg[i]>=14||near.has(i))&&scale>0.22){ctx.globalAlpha=dim?.2:1;ctx.fillStyle='#e6edf3';ctx.font='11px ui-sans-serif,system-ui';ctx.fillText(D.label[i],gx+r+3,y+3);}
  ctx.globalAlpha=1;}
function loop(){if(alive)for(let k=0;k<3;k++)step();if(!fitted&&iter>30){fitVisible();fitted=true;}draw();requestAnimationFrame(loop);}
let panning=false,lastX=0,lastY=0;
function pick(sx,sy){let best=-1,bd=1e9;for(let i=0;i<N;i++){if(!visible(i))continue;const dx=SX(px[i])-sx,dy=SY(py[i])-sy,dd=dx*dx+dy*dy,rr=(radius(i)*Math.max(.6,Math.min(1.8,scale))+4)**2;if(dd<rr&&dd<bd){bd=dd;best=i;}}return best;}
cv.addEventListener('mousemove',ev=>{const sx=ev.clientX,sy=ev.clientY;if(dragNode>=0){px[dragNode]=WX(sx);py[dragNode]=WY(sy);temp=Math.max(temp,K*1.4);alive=true;return;}if(panning){ox+=sx-lastX;oy+=sy-lastY;lastX=sx;lastY=sy;return;}hover=pick(sx,sy);const tip=document.getElementById('tip');if(hover>=0){tip.style.opacity=1;tip.style.left=(sx+12)+'px';tip.style.top=(sy+12)+'px';tip.textContent=D.label[hover]+' · '+D.kind[hover]+' · '+file[hover]+':'+D.line[hover];}else tip.style.opacity=0;});
cv.addEventListener('mousedown',ev=>{const h=pick(ev.clientX,ev.clientY);if(h>=0)dragNode=h;else{panning=true;lastX=ev.clientX;lastY=ev.clientY;}});
addEventListener('mouseup',()=>{dragNode=-1;panning=false;});
cv.addEventListener('wheel',ev=>{ev.preventDefault();const f=ev.deltaY<0?1.12:1/1.12,wx=WX(ev.clientX),wy=WY(ev.clientY);scale*=f;ox=ev.clientX-wx*scale;oy=ev.clientY-wy*scale;},{passive:false});
cv.addEventListener('dblclick',fitVisible);
document.getElementById('fit').onclick=fitVisible;
addEventListener('keydown',e=>{if((e.key==='f'||e.key==='F')&&!/INPUT|TEXTAREA/.test(document.activeElement.tagName))fitVisible();});
// panel
document.getElementById('nc').textContent=N;document.getElementById('ec').textContent=E.length;document.getElementById('sc').textContent=Object.keys(bySector).length;
const sizes=[0,0,0];for(let i=0;i<N;i++)sizes[cat[i]]++;
[['Client code','dot','#c9d1d9'],['Tests','dot','#c9d1d9'],['External libs','ring','']].forEach((cd,ci)=>{const l=document.createElement('label');l.className='row';const cb=document.createElement('input');cb.type='checkbox';cb.checked=true;cb.onchange=()=>catOn[ci]=cb.checked;const sw=document.createElement('span');sw.className=cd[1];if(cd[2])sw.style.background=cd[2];const t=document.createElement('span');t.textContent=cd[0];const c=document.createElement('span');c.className='cnt';c.textContent=sizes[ci];l.append(cb,sw,t,c);document.getElementById('cats').append(l);});
document.getElementById('zones').onchange=e=>showZones=e.target.checked;
document.getElementById('d3').onchange=e=>threeD=e.target.checked;
const root={name:'',children:{},count:0,path:''};
for(let i=0;i<N;i++){let node=root,path='';const parts=file[i].split('/');for(let p=0;p<parts.length;p++){path=path?path+'/'+parts[p]:parts[p];if(!node.children[parts[p]])node.children[parts[p]]={name:parts[p],children:{},count:0,path:path,leaf:p===parts.length-1};node=node.children[parts[p]];node.count++;}}
const treeEl=document.getElementById('tree');
const expanded=new Set();(function ce(n){Object.values(n.children).forEach(ch=>{if(Object.keys(ch.children).length){expanded.add(ch.path);ce(ch);}});})(root); // default: all dirs open
function renderTree(){treeEl.innerHTML='';(function walk(n,depth){Object.values(n.children).sort((a,b)=>b.count-a.count).forEach(ch=>{treeEl.append(rowEl(ch,depth));if(Object.keys(ch.children).length&&expanded.has(ch.path))walk(ch,depth+1);});})(root,0);}
function rowEl(node,depth){const l=document.createElement('div');l.className='trow';l.style.paddingLeft=(depth*13)+'px';const has=Object.keys(node.children).length>0;const toggle=()=>{expanded.has(node.path)?expanded.delete(node.path):expanded.add(node.path);renderTree();};const tw=document.createElement('span');tw.className='tw';tw.textContent=has?(expanded.has(node.path)?'▾':'▸'):'';if(has)tw.onclick=toggle;const cb=document.createElement('input');cb.type='checkbox';cb.checked=!isHidden(node.path);cb.onchange=()=>{cb.checked?hidden.delete(node.path):hidden.add(node.path);renderTree();};const dot=document.createElement('span');dot.className='dot';dot.style.background=colorFor(node.leaf?dirOf(node.path):node.path);const t=document.createElement('span');t.className='tn';t.textContent=node.name;if(has){t.style.cursor='pointer';t.onclick=toggle;}const c=document.createElement('span');c.className='cnt';c.textContent=node.count;l.append(tw,cb,dot,t,c);return l;}
document.getElementById('exp').onclick=()=>{(function all(n){Object.values(n.children).forEach(ch=>{if(Object.keys(ch.children).length){expanded.add(ch.path);all(ch);}});})(root);renderTree();};
document.getElementById('col').onclick=()=>{expanded.clear();renderTree();};renderTree();
const panel=document.getElementById('panel'),rz=document.getElementById('resizer');let rzx=0,rzw=0,rzon=false;
rz.addEventListener('mousedown',ev=>{rzon=true;rzx=ev.clientX;rzw=panel.offsetWidth;ev.preventDefault();});
addEventListener('mousemove',ev=>{if(rzon)panel.style.width=Math.max(210,Math.min(680,rzw+(ev.clientX-rzx)))+'px';});
addEventListener('mouseup',()=>rzon=false);loop();

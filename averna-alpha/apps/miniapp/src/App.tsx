import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import { ArrowDownLeft, ArrowUpRight, BarChart3, Boxes, Check, ChevronRight, CircleUserRound, LayoutDashboard, LogOut, Package, Plus, Search, ShoppingBag, ShoppingCart, Store, X } from 'lucide-react'

declare global { interface Window { Telegram?: { WebApp?: { ready: () => void; expand: () => void; themeParams?: Record<string,string>; initData?: string; colorScheme?: string } } } }
type Product = { id: string; name: string; description: string; price_minor: number; currency: string; is_active: boolean }
type Org = { id: string; name: string; slug: string; role: string }
type Order = { id: string; status: string; total_minor: number; currency: string; created_at: string; customer_name: string; phone: string; email: string }
const API = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/$/, '')
const money = (minor: number, currency = 'EUR') => new Intl.NumberFormat('ru-RU', { style: 'currency', currency }).format(minor / 100)

export default function App() {
  const [mode, setMode] = useState<'shop'|'business'>('shop')
  const [products, setProducts] = useState<Product[]>([])
  const [cart, setCart] = useState<Record<string,number>>({})
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(false)
  const [notice, setNotice] = useState('')
  const [customer, setCustomer] = useState({ name: '', phone: '', email: '' })
  const [checkoutOpen, setCheckoutOpen] = useState(false)
  const [token, setToken] = useState(sessionStorage.getItem('averna_access') || '')
  const [refresh, setRefresh] = useState(sessionStorage.getItem('averna_refresh') || '')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [registerMode, setRegisterMode] = useState(false)
  const [displayName, setDisplayName] = useState('')
  const [orgName, setOrgName] = useState('')
  const [orgSlug, setOrgSlug] = useState('')
  const [orgs, setOrgs] = useState<Org[]>([])
  const [orgID, setOrgID] = useState('')
  const [orders, setOrders] = useState<Order[]>([])
  const [newProduct, setNewProduct] = useState({ name: '', description: '', price: '' })
  const [storeSlug, setStoreSlug] = useState(new URLSearchParams(location.search).get('store') || 'demo')
  const activeOrg = orgs.find(o => o.id === orgID)
  const cartCount = Object.values(cart).reduce((a,b) => a+b, 0)
  const total = useMemo(() => products.reduce((sum,p) => sum + (cart[p.id] || 0) * p.price_minor, 0), [products,cart])

  useEffect(() => { window.Telegram?.WebApp?.ready(); window.Telegram?.WebApp?.expand() }, [])
  useEffect(() => { if (mode === 'shop') loadCatalog() }, [mode, storeSlug])
  useEffect(() => { if (mode === 'business' && token && orgID) loadBusiness() }, [mode, token, orgID])
  useEffect(() => { if (!token) return; fetch(`${API}/v1/organizations`,{headers:{Authorization:`Bearer ${token}`}}).then(r=>{if(!r.ok)throw new Error('Session expired');return r.json()}).then(data=>{if(data?.items){setOrgs(data.items);if(data.items[0])setOrgID(current=>current||data.items[0].id)}}).catch(()=>{setToken('');sessionStorage.removeItem('averna_access');sessionStorage.removeItem('averna_refresh')}) }, [])
  async function request(path: string, init: RequestInit = {}) {
    const headers = new Headers(init.headers)
    if (init.body) headers.set('Content-Type','application/json')
    if (token) headers.set('Authorization',`Bearer ${token}`)
    const response = await fetch(`${API}${path}`,{...init,headers})
    if (!response.ok) { const body = await response.json().catch(()=>null); throw new Error(body?.error?.message || `Ошибка запроса (${response.status})`) }
    return response.status === 204 ? null : response.json()
  }
  async function loadCatalog() {
    setLoading(true)
    try { const data = await request(`/v1/store/${encodeURIComponent(storeSlug)}/products`); setProducts(data.items || []) }
    catch (e) { setNotice(e instanceof Error ? e.message : 'Не удалось загрузить каталог') }
    finally { setLoading(false) }
  }
  async function login(e: FormEvent) {
    e.preventDefault(); setNotice('')
    try {
      const path = registerMode ? '/v1/auth/register' : '/v1/auth/login'
      const payload = registerMode
        ? { email, password, display_name: displayName, organization_name: orgName, organization_slug: orgSlug }
        : { email, password }
      const data = await request(path,{method:'POST',body:JSON.stringify(payload)})
      setToken(data.tokens.access_token); setRefresh(data.tokens.refresh_token)
      sessionStorage.setItem('averna_access',data.tokens.access_token); sessionStorage.setItem('averna_refresh',data.tokens.refresh_token)
      const orgData = await fetch(`${API}/v1/organizations`,{headers:{Authorization:`Bearer ${data.tokens.access_token}`}}).then(r=>r.json())
      setOrgs(orgData.items || []); if (orgData.items?.[0]) setOrgID(orgData.items[0].id); setNotice(registerMode ? 'Бизнес-пространство создано' : 'Вход выполнен')
    } catch (e) { setNotice(e instanceof Error ? e.message : 'Не удалось выполнить вход') }
  }
  async function loadBusiness() {
    if (!orgID) return
    try {
      const [p,o] = await Promise.all([request(`/v1/organizations/${orgID}/products`),request(`/v1/organizations/${orgID}/orders`)])
      setProducts(p.items || []); setOrders(o.items || [])
    } catch(e) { setNotice(e instanceof Error ? e.message : 'Не удалось загрузить данные') }
  }
  async function addProduct(e: FormEvent) {
    e.preventDefault(); if (!orgID) return
    try { await request(`/v1/organizations/${orgID}/products`,{method:'POST',body:JSON.stringify({name:newProduct.name,description:newProduct.description,price_minor:Math.round(Number(newProduct.price)*100)})}); setNewProduct({name:'',description:'',price:''}); await loadBusiness(); setNotice('Товар добавлен') }
    catch(e) { setNotice(e instanceof Error ? e.message : 'Не удалось добавить товар') }
  }
  async function toggleProduct(product: Product) {
    if (!orgID) return
    try {
      await request(`/v1/organizations/${orgID}/products/${product.id}`, { method: 'PATCH', body: JSON.stringify({ name: product.name, description: product.description, price_minor: product.price_minor, is_active: !product.is_active }) })
      await loadBusiness()
      setNotice(product.is_active ? 'Товар скрыт из витрины' : 'Товар снова опубликован')
    } catch (e) { setNotice(e instanceof Error ? e.message : 'Не удалось обновить товар') }
  }
  async function changeOrderStatus(order: Order, status: 'confirmed'|'completed'|'cancelled') {
    if (!orgID) return
    try {
      await request(`/v1/organizations/${orgID}/orders/${order.id}`, { method: 'PATCH', body: JSON.stringify({ status }) })
      await loadBusiness()
      setNotice(`Статус заказа обновлён: ${status}`)
    } catch (e) { setNotice(e instanceof Error ? e.message : 'Не удалось обновить заказ') }
  }
  async function checkout(e: FormEvent) {
    e.preventDefault()
    try {
      const items = Object.entries(cart).filter(([,q])=>q>0).map(([product_id,quantity])=>({product_id,quantity}))
      const data = await request(`/v1/store/${encodeURIComponent(storeSlug)}/orders`,{method:'POST',headers:{'Idempotency-Key':crypto.randomUUID()},body:JSON.stringify({customer_name:customer.name,phone:customer.phone,email:customer.email,items})})
      setCart({}); setCheckoutOpen(false); setNotice(`Заказ создан · ${money(data.total_minor,data.currency)} · № ${data.order_id.slice(0,8)}`)
    } catch(e) { setNotice(e instanceof Error ? e.message : 'Не удалось оформить заказ') }
  }
  function logout() { setToken('');setRefresh('');setOrgs([]);setOrders([]);sessionStorage.removeItem('averna_access');sessionStorage.removeItem('averna_refresh') }
  const filtered = products.filter(p => p.name.toLowerCase().includes(query.toLowerCase()))

  return <main className="shell">
    <header className="topbar"><a className="brand" href="#"><span className="brand-mark">A</span><span>AVERNA<small>ALPHA · TELEGRAM COMMERCE</small></span></a><div className="top-actions"><span className="status-dot"/> <span className="live-label">В сети</span><button className="avatar"><CircleUserRound size={19}/></button></div></header>
    <section className="hero"><div className="hero-copy"><span className="eyebrow">ОДНО ПРОСТРАНСТВО ДЛЯ БИЗНЕСА</span><h1>{mode==='shop'?'Покупки без лишних шагов.':'Ваш бизнес — под контролем.'}</h1><p>{mode==='shop'?'Откройте каталог, выберите товары и оформите заказ прямо в Telegram.':'Управляйте товарами и заказами в одном удобном рабочем пространстве.'}</p></div><div className="hero-orb"><ShoppingBag size={34}/><span/></div></section>
    <nav className="mode-switch"><button className={mode==='shop'?'active':''} onClick={()=>setMode('shop')}><Store size={17}/> Магазин</button><button className={mode==='business'?'active':''} onClick={()=>{setMode('business');if(token)loadBusiness()}}><LayoutDashboard size={17}/> Управление бизнесом</button></nav>
    {notice && <div className="notice" role="status">{notice}<button onClick={()=>setNotice('')}><X size={15}/></button></div>}
    {mode==='shop' ? <section className="content"><div className="section-head"><div><span className="eyebrow">ВИТРИНА</span><h2>Популярные товары</h2></div><span className="count-pill">{products.length} позиций</span></div><div className="search"><Search size={18}/><input value={query} onChange={e=>setQuery(e.target.value)} placeholder="Найти товар..."/></div>{loading?<div className="empty">Загружаем каталог…</div>:filtered.length===0?<div className="empty"><Package size={25}/><b>Пока нет товаров</b><span>Когда магазин добавит товары, они появятся здесь.</span></div>:<div className="product-grid">{filtered.map((p,i)=><article className="product-card" key={p.id}><div className={`product-art art-${i%4}`}><Package size={32}/><span>AVERNA GOODS</span></div><div className="product-info"><h3>{p.name}</h3><p>{p.description || 'Качественный товар от магазина.'}</p><div className="product-bottom"><strong>{money(p.price_minor,p.currency)}</strong><button className="add-btn" onClick={()=>setCart(c=>({...c,[p.id]:(c[p.id]||0)+1}))}><Plus size={17}/></button></div>{cart[p.id]>0&&<div className="quantity"><button onClick={()=>setCart(c=>({...c,[p.id]:Math.max(0,(c[p.id]||0)-1)}))}>−</button><span>{cart[p.id]} в корзине</span><button onClick={()=>setCart(c=>({...c,[p.id]:(c[p.id]||0)+1}))}>+</button></div>}</div></article>)}</div>}
      <div className="cart-bar"><div className="cart-icon"><ShoppingCart size={19}/><i>{cartCount}</i></div><div className="cart-summary"><b>{cartCount} товара</b><span>Итого: {money(total,products[0]?.currency || 'EUR')}</span></div><button className="primary" disabled={!cartCount} onClick={()=>setCheckoutOpen(true)}>Оформить <ChevronRight size={17}/></button></div>
    </section> : <section className="content">{!token?<div className="login-card"><div className="login-icon"><LayoutDashboard size={24}/></div><h2>{registerMode?'Создать бизнес':'Вход для владельца'}</h2><p>{registerMode?'Создайте аккаунт и своё бизнес-пространство AVERNA.':'Авторизуйтесь, чтобы управлять товарами и заказами вашей организации.'}</p><form onSubmit={login}>{registerMode&&<><label>Ваше имя<input value={displayName} onChange={e=>setDisplayName(e.target.value)} required placeholder="Имя"/></label><label>Название бизнеса<input value={orgName} onChange={e=>setOrgName(e.target.value)} required placeholder="Например, North Store"/></label><label>Адрес магазина (латиницей)<input value={orgSlug} onChange={e=>setOrgSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g,''))} required minLength={3} maxLength={48} placeholder="north-store"/></label></>}<label>Email<input type="email" value={email} onChange={e=>setEmail(e.target.value)} required placeholder="you@example.com"/></label><label>Пароль<input type="password" value={password} onChange={e=>setPassword(e.target.value)} required minLength={12} placeholder="Минимум 12 символов"/></label><button className="primary full">{registerMode?'Создать пространство':'Войти'} <ArrowUpRight size={17}/></button></form><button className="text-switch" onClick={()=>setRegisterMode(v=>!v)}>{registerMode?'Уже есть аккаунт? Войти':'Впервые в AVERNA? Создать бизнес'}</button></div>:<><div className="section-head"><div><span className="eyebrow">РАБОЧЕЕ ПРОСТРАНСТВО</span><h2>Обзор бизнеса</h2></div><button className="ghost" onClick={logout}><LogOut size={16}/> Выйти</button></div><label className="org-select">Организация<select value={orgID} onChange={e=>setOrgID(e.target.value)}>{orgs.map(o=><option key={o.id} value={o.id}>{o.name} · {o.role}</option>)}</select></label>{activeOrg&&<div className="store-link"><div><span className="eyebrow">ССЫЛКА НА ВИТРИНУ</span><b>/{activeOrg.slug}</b><small>{window.location.origin}/?store={activeOrg.slug}</small></div><button className="secondary" onClick={()=>{setStoreSlug(activeOrg.slug);setMode('shop')}}><Store size={15}/> Открыть</button></div>}<div className="stats"><div className="stat-card"><span>Заказы</span><b>{orders.length}</b><small><ArrowUpRight size={13}/> Последние 100</small></div><div className="stat-card"><span>Товары</span><b>{products.length}</b><small><Boxes size={13}/> В каталоге</small></div><div className="stat-card"><span>Выручка*</span><b>{money(orders.reduce((s,o)=>s+o.total_minor,0),orders[0]?.currency||'EUR')}</b><small><BarChart3 size={13}/> По загруженным заказам</small></div></div><div className="manager-grid"><div className="panel"><div className="panel-head"><h3><Package size={17}/> Товары</h3><span>{products.length}</span></div><form className="product-form" onSubmit={addProduct}><input value={newProduct.name} onChange={e=>setNewProduct(v=>({...v,name:e.target.value}))} placeholder="Название товара" required/><input value={newProduct.description} onChange={e=>setNewProduct(v=>({...v,description:e.target.value}))} placeholder="Описание"/><input type="number" min="0" step="0.01" value={newProduct.price} onChange={e=>setNewProduct(v=>({...v,price:e.target.value}))} placeholder="Цена в EUR" required/><button className="primary"><Plus size={16}/> Добавить товар</button></form><div className="compact-list">{products.map(p=><div className="compact-row" key={p.id}><div className="mini-product"><Package size={16}/></div><div><b>{p.name}</b><small>{p.is_active?'Опубликован':'Скрыт из витрины'}</small></div><strong>{money(p.price_minor,p.currency)}</strong><button className="row-action" onClick={()=>toggleProduct(p)}>{p.is_active?'Скрыть':'Показать'}</button></div>)}</div></div><div className="panel"><div className="panel-head"><h3><ShoppingBag size={17}/> Последние заказы</h3><span>{orders.length}</span></div>{orders.length===0?<div className="empty small-empty">Заказов пока нет</div>:<div className="compact-list">{orders.map(o=><div className="compact-row order-row" key={o.id}><div className="order-icon"><ShoppingBag size={16}/></div><div><b>{o.customer_name || `Заказ #${o.id.slice(0,8)}`}</b><small>{o.phone || o.email || new Date(o.created_at).toLocaleDateString('ru-RU')}</small><small className="status-label">{({pending:'Ожидает',confirmed:'Подтверждён',completed:'Завершён',cancelled:'Отменён'} as Record<string,string>)[o.status] || o.status}</small></div><div className="order-end"><strong>{money(o.total_minor,o.currency)}</strong><div className="row-actions">{o.status==='pending'&&<><button className="row-action" onClick={()=>changeOrderStatus(o,'confirmed')}>Принять</button><button className="row-action danger" onClick={()=>changeOrderStatus(o,'cancelled')}>Отменить</button></>}{o.status==='confirmed'&&<><button className="row-action" onClick={()=>changeOrderStatus(o,'completed')}>Завершить</button><button className="row-action danger" onClick={()=>changeOrderStatus(o,'cancelled')}>Отменить</button></>}</div></div></div>)}</div>}<button className="secondary full" onClick={loadBusiness}>Обновить данные <ArrowDownLeft size={16}/></button></div></div><p className="footnote">* Сумма отображает только последние 100 заказов, а не бухгалтерскую выручку.</p></>}</section>}
    <footer><span>© 2026 AVERNA ALPHA</span><span>Бизнес и покупки — в одном месте</span></footer>
    {checkoutOpen&&<div className="modal-backdrop" onClick={()=>setCheckoutOpen(false)}><form className="checkout-modal" onSubmit={checkout} onClick={e=>e.stopPropagation()}><div className="modal-head"><div><span className="eyebrow">ПОСЛЕДНИЙ ШАГ</span><h2>Оформление заказа</h2></div><button type="button" className="icon-btn" onClick={()=>setCheckoutOpen(false)}><X size={18}/></button></div><div className="checkout-lines">{products.filter(p=>cart[p.id]>0).map(p=><div key={p.id}><span>{p.name} × {cart[p.id]}</span><b>{money(p.price_minor*cart[p.id],p.currency)}</b></div>)}</div><label>Ваше имя<input value={customer.name} onChange={e=>setCustomer(v=>({...v,name:e.target.value}))} required placeholder="Как к вам обращаться?"/></label><label>Телефон<input value={customer.phone} onChange={e=>setCustomer(v=>({...v,phone:e.target.value}))} required placeholder="+32 ..."/></label><label>Email <small>необязательно</small><input type="email" value={customer.email} onChange={e=>setCustomer(v=>({...v,email:e.target.value}))} placeholder="you@example.com"/></label><div className="checkout-total"><span>Итого</span><b>{money(total,products[0]?.currency||'EUR')}</b></div><button className="primary full">Подтвердить заказ <Check size={17}/></button><p className="footnote">Оплата пока не подключена. После оформления магазин получит заказ и свяжется с вами.</p></form></div>}
  </main>
}

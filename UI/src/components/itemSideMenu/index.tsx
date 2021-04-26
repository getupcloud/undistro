import React, { FC, useState } from 'react'
import Classnames from 'classnames'

import './index.scss'

type SubItem = {
  name: string
}

type Item = {
  name: string
  icon: string
  // subitens: SubItem[]
}

const ItemSideMenu: FC<Item> = ({
	name,
	icon
}) => {
	const [show, setShow] = useState<boolean>(false)
	const style = Classnames('side-item', {
		active: show
	})

	return (
		<>
			<li onClick={() => setShow(!show)} className={style}>
				<i className={icon} />
				<p>{name}</p>
				{show ? <i className='icon-arrow-up' /> : <i className='icon-arrow-down' />}
			</li>
			{show && <div className='item-menu'>
				<p>Subitem 01</p>
				<p>Subitem 02</p>
				<p>Subitem 03</p>
				<p>Subitem 04</p>
				<p>Subitem 05</p>
			</div>}
		</>
	)
}

export default ItemSideMenu
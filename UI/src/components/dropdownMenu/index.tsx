import React, { FC } from 'react'

import './index.scss'

type Item = {
  name: string
}

const DropdownMenu: FC = () => {
	return (
		<div className='dropdown-menu-container'>
			<ul className='itens'>
				<li className='item'>Item 1</li>
				<li className='item'>Item 1</li>
				<li className='item'>Item 1</li>
				<li className='item'>Item 1</li>
			</ul>
		</div>
	)
}

export default DropdownMenu
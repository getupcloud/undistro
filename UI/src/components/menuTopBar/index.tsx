import React, { FC } from 'react'
import Dropdown from '@components/dropdownMenu'
import Logo from '@assets/images/logo.png'
import './index.scss'

const MenuTopBar: FC = () => {
	return (
		<div className='menu-top-container'>
			<div className='img-container'>
				<img alt='UnDistro logo' src={Logo} />
			</div>
			<ul>
				<li>
					<p>Create</p>
					<Dropdown />
				</li>
				<li>
					<p>Manage</p>
					<Dropdown />
				</li>
				<li>
					<p>Settings</p>
					<Dropdown />
				</li>
				<li>
					<p>User</p>
					<Dropdown />
				</li>
				<li>
					<p>Options</p>
					<Dropdown />
				</li>
			</ul>
		</div>
	)
}

export default MenuTopBar
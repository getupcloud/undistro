import React, { FC } from 'react'
// import Dropdown from '@components/dropdownMenu'
import Logo from '@assets/images/logo.png'
import './index.scss'

const MenuTopBar: FC = () => {
	return (
		<div className='menu-top-container'>
			<div className='img-container'>
				<img alt='undistro-logo' src={Logo} />
			</div>
			<ul>
				<li>
					<p>Create</p>
				</li>
				<li>
					<p>Manage</p>
				</li>
				<li>
					<p>Preferences</p>
				</li>
			</ul>
		</div>
	)
}

export default MenuTopBar
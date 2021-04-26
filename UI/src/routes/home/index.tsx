import React from 'react'
import MenuTop from '@components/menuTopBar'
import MenuSide from '@components/menuSideBar'
import './index.scss'

export default function HomePage () {
	return (
		<div className='home-page-route'>
			<MenuTop />
			<MenuSide />
		</div>
	)
}
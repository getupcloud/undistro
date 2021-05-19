import React from 'react'
import { Switch, Route } from 'react-router-dom'
import HomePageRoute from '@routes/home'
import TestRoute from '@routes/test'
import MenuTop from '@components/menuTopBar'
import MenuSide from '@components/menuSideBar'
import Modals from './modals'
import 'styles/app.scss'

export default function App() {
	return (
		<div className='route-container'>
			<MenuTop />
			<MenuSide />

			<div className='route-content'>
				<Switch>
						<Route path='/' component={HomePageRoute} />
						<Route exact path='/test' component={TestRoute} />
				</Switch>
				<Modals />
				</div>
		</div>
	)
}

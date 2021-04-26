import React from 'react'
import { Switch, Route } from 'react-router-dom'
import HomePageRoute from '@routes/home'
import TestRoute from '@routes/test'

import './index.css'

export default function App() {
	return (
		<>
			<Switch>
				<Route path='/' component={HomePageRoute} />
				<Route exact path='/test' component={TestRoute} />
			</Switch>
		</>
	)
}

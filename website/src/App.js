import React from 'react'
import { Switch, Route, useHistory } from 'react-router-dom'
import HomePageRoute from 'Routes/home'
import FaqRoute from 'Routes/faq'
import DocsRoute from 'Routes/docs'
import CommunityRoute from 'Routes/community'
import Navbar from 'Components/atoms/navbar'
import Footer from 'Components/atoms/footer'
import Logo from 'Assets/images/logo.png'

export default function App () {
	const history = useHistory()

	const redirectHome = () => {
		return history.push('/')
	}

	return (
		<>
			<Navbar onClick={redirectHome} white className='undistro-logo' img={Logo} />
			<Switch>
				<Route exact path='/' component={HomePageRoute} />
				<Route path='/docs' component={DocsRoute}/>
				<Route path='/faq' component={FaqRoute} />
				<Route path='/community' component={CommunityRoute} />
				<Route component={HomePageRoute} />
			</Switch>
			<Footer />
		</>
	)
}

import React from 'react'
import ReactDOM from 'react-dom'
import ReactGA from 'react-ga'
import { createBrowserHistory } from 'history'
import { Router, Route } from 'react-router-dom'
import { QueryParamProvider } from 'use-query-params'
import App from './App'
import { ScrollTopRouter } from 'Util/helpers'
import 'Styles/index.scss'
import 'Assets/font-icon/icons.css'

const app = document.getElementById('app')
ReactGA.initialize('UA-193342188-1')
const history = createBrowserHistory()
// Initialize google analytics page view tracking
history.listen(location => {
	ReactGA.set({ page: location.pathname }) // Update the user's current page
	ReactGA.pageview(location.pathname) // Record a pageview for the given page
})

const main = (
	<Router history={history}>
		<QueryParamProvider ReactRouterRoute={Route}>
			<ScrollTopRouter />
			<App />
		</QueryParamProvider>
	</Router>
)

ReactDOM.render(main, app)

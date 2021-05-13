import React from 'react'
import ReactDOM from 'react-dom'
import App from './App'
import reportWebVitals from './reportWebVitals'
import { BrowserRouter as Router } from 'react-router-dom'
// import { QueryParamProvider } from 'use-query-params'

import 'styles/app.scss'
import '@assets/font-icon/icons.css'

ReactDOM.render(
	<Router>
		{/* <QueryParamProvider ReactRouterRoute={Route}> */}
		<App />
		{/* </QueryParamProvider> */}
	</Router>,
	document.getElementById('root')
)

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals()

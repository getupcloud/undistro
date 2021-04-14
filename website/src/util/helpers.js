import { useEffect } from 'react'
import { useLocation } from 'react-router-dom'

export const scrollTop = (mode) => {
	document.getElementById('app').scrollIntoView({ behavior: mode || 'auto' })
}

export const ScrollTopRouter = () => {
	const { pathname } = useLocation()

	useEffect(() => {
		window.scrollTo(0, 0)
	}, [pathname])

	return null
}

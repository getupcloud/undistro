const path = require('path')

exports.onCreateWebpackConfig = (args) => {
	args.actions.setWebpackConfig({
		resolve: {
			modules: [path.resolve(__dirname, '../src'), 'node_modules'],
			alias: {
				Components: path.resolve(__dirname, '../src/components/'),
				Assets: path.resolve(__dirname, '../src/assets/'),
				Util: path.resolve(__dirname, '../src/util/'),
				Routes: path.resolve(__dirname, '../src/routes/'),
				Styles: path.resolve(__dirname, '../src/styles/')
			}
		}
	})
}

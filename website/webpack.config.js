// const Dotenv = require('dotenv-webpack')
const path = require('path')
const slug = require('rehype-slug')
const highlight = require('remark-highlight.js')

const HtmlWebPackPlugin = require('html-webpack-plugin')
const htmlPlugin = new HtmlWebPackPlugin({
	template: './public/index.html',
	favicon: './public/favicon.png',
	filename: './index.html'
})

module.exports = {
	entry: ['@babel/polyfill', './src'],
	output: {
		path: path.resolve(__dirname, './dist'),
		filename: 'bundle.js',
		publicPath: '/'
	},
	devServer: {
		historyApiFallback: {
			disableDotRule: true
		},
		inline: true,
		host: '0.0.0.0',
		port: 3001
	},
	resolve: {
		alias: {
			Components: path.resolve(__dirname, 'src/components/'),
			Assets: path.resolve(__dirname, 'src/assets/'),
			Util: path.resolve(__dirname, 'src/util/'),
			Routes: path.resolve(__dirname, 'src/routes/'),
			Styles: path.resolve(__dirname, 'src/styles/')
		}
	},
	module: {
		rules: [
			{
				test: /\.mdx?$/,
				use: [
					{
						loader: 'babel-loader'
					},
					{
						loader: '@mdx-js/loader',
						options: {
							remarkPlugins: [highlight],
							rehypePlugins: [slug]
						}
					}
				]
			},
			{
				test: /\.js$/,
				exclude: /node_modulest/,
				use: {
					loader: 'babel-loader',
					options: {
						presets: [['@babel/preset-env', {
							targets: {
								browsers: ['last 2 versions', 'ie >= 11']
							},
							useBuiltIns: 'entry'
						}], '@babel/preset-react']
					}
				}
			},
			{
				test: /\.(scss|css)$/,
				exclude: /node_modules/,
				use: [
					{ loader: 'style-loader' },
					{ loader: 'css-loader' },
					{ loader: 'sass-loader' },
					{ loader: 'postcss-loader' }
				]
			},
			{
				test: /\.(ttf|eot|icons\.svg|woff|woff2)(\?v=[0-9]\.[0-9]\.[0-9])?$/,
				loader: 'file-loader'
			},
			{
				test: /\.(gif|png|jpe?g|svg)$/i,
				use: [
					{ loader: 'file-loader' },
					{
						loader: 'image-webpack-loader',
						options: {
							gifsicle: {
								interlanced: false
							},
							optipng: {
								optimizationLevel: 7
							},
							pngquant: {
								quality: [0.65, 0.90],
								speed: 4
							},
							mozjpeg: {
								progressive: true,
								quality: 65
							}
						}
					}
				]
			}
		]
	},
	plugins: [
		htmlPlugin
	]
}

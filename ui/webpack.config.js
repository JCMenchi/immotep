const path = require("path");
const webpack = require("webpack");
const { CleanWebpackPlugin } = require('clean-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const CopyPlugin = require("copy-webpack-plugin");

module.exports = {
  entry: './frontend/index.js',
  mode: "development",
  devtool: "source-map",
  module: {
    noParse: /jspdf.umd.min.js/,
    rules: [
      {
        test: /\-worker\.js$/,
        use: { loader: 'worker-loader',
        options: { inline:"fallback" },}
      },
      {
        test: /\.(js|jsx)$/,
        exclude: /(node_modules|bower_components)/,
        loader: "babel-loader",
        options: { presets: ["@babel/env"] }
      },
      {
        test: /\.css$/,
        use: ["style-loader", "css-loader"]
      },
      {
        test: /\.(png|jpg|jpeg|gif|svg|woff|woff2)$/,
        loader: "file-loader"
      }
    ]
  },
  resolve: { extensions: ["*", ".js", ".jsx"] },
  output: {
    path: path.resolve(__dirname, "immotep/"),
    publicPath: "/immotep/",
    filename: "[name].bundle.js"
  },
  optimization: {
    splitChunks: {
      chunks: "all",
      minSize: 600*1024,
    },
  },
  devServer: {
    
    historyApiFallback: {
      index: '/immotep/index.html'
    },
    port: 3000,
    static: {
      directory: path.join(__dirname, "frontend/"),
      publicPath: "http://localhost:3000/immotep",
    },
    hot: true,
    proxy: [
      {
        context: ['/api'],
        target: 'http://34.44.29.28/',
      },
    ]
  },
  plugins: [
    new webpack.ProgressPlugin(),
    new CleanWebpackPlugin({verbose:true}), 
    new webpack.HotModuleReplacementPlugin(),
    new HtmlWebpackPlugin({
      minify: false,
      title: 'immotep',
      template: './frontend/index.tpl',
      meta: {
        'viewport': 'initial-scale=1, width=device-width',
        'mobile-web-app-capable': 'yes'
      },
      base: '/immotep',
      favicon: 'frontend/favicon.ico'
    }),
    new CopyPlugin({
      patterns: [
        { from: "frontend/manifest.json" },
        { from: "frontend/locales", to: "locales" }
      ]
    })
  ]
};

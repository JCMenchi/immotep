import axios from 'axios';

import log from 'loglevel';

const LOGGER = log.getLogger("api");
LOGGER.setDefaultLevel("debug");

/*==============================================================*/
/**            Setup AXIOS instance                             */
/*==============================================================*/
const axiosInstance = axios.create({
    timeout: 15000,
    headers: { 'content-type': 'application/json' }
});

export default axiosInstance;

// Add a request interceptor
axiosInstance.interceptors.request.use(function (config) {
    LOGGER.debug(`Send (${config.method}) request to ${config.url}`);
    return config;
}, function (error) {
    if (error.request) {
        LOGGER.debug(`Error(${error.request.status}) in axios request to (${error.config.method})${error.config.url}: ${error.request.data}`);
    } else {
        LOGGER.debug(`Error ${error.message}`);
    }
    return Promise.reject(error);
});

// Add a response interceptor
axiosInstance.interceptors.response.use(function (response) {
    LOGGER.debug(`Get response of (${response.config.method}) request to ${response.config.url}:`, response.data);
    return response;
}, function (error) {
    if (error.response) {
        // Any status codes that falls outside the range of 2xx cause this function to trigger
        LOGGER.debug(`Error(${error.response.status}) in axios response to (${error.config.method})${error.config.url}: ${error.response.data}`);
    } else if (error.config) {
        LOGGER.debug(`Error in axios response to (${error.config.method})${error.config.url}: ${error.message}`);
    } else {
        LOGGER.debug(`Error ${error.message}`);
    }
    return Promise.reject(error);
});

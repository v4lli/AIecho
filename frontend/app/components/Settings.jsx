import {useEffect, useState} from 'react';
import {Desktop} from "@/app/components/Responsive";

export const Settings = ({onVideoDeviceChange}) => {
    const [videoDevices, setVideoDevices] = useState([]);
    const [selectedVideoDeviceID, setSelectedVideoDeviceID] = useState('');

    useEffect(() => {
        const getDevices = async () => {
            try {
                await navigator.mediaDevices.getUserMedia({audio: false, video: true});
                const deviceInfos = await navigator.mediaDevices.enumerateDevices();
                const videoInputs = deviceInfos.filter(device => device.kind === 'videoinput');
                setVideoDevices(videoInputs);
                if (videoInputs.length > 0) setSelectedVideoDeviceID(videoInputs[0].deviceId);
            } catch (err) {
                console.error('Error fetching devices:', err);
            }
        };
        getDevices();
    }, []);

    useEffect(() => {
        if (selectedVideoDeviceID) {
            onVideoDeviceChange(selectedVideoDeviceID);
        }
    }, [selectedVideoDeviceID, onVideoDeviceChange]);


    return (<div className="settings">
        <div className="settings-device" aria-label="Video Input Settings">
            <label htmlFor={'videoInput'}>Video Input: </label>
            <select onChange={(e) => setSelectedVideoDeviceID(e.target.value)} value={selectedVideoDeviceID}
                    id={'videoInput'} aria-label={'Select Video Input Device'}>
                {videoDevices.map((videoDevice) => (<option key={videoDevice.deviceId} value={videoDevice.deviceId}>
                    {videoDevice.label}
                </option>))}
            </select>
        </div>
    </div>);
};